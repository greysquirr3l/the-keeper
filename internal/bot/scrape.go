// File: internal/bot/scrape.go

package bot

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Updated version of StartPeriodicScraping to only notify if new codes are found
func (b *Bot) StartPeriodicScraping() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Adjust the interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				results, err := b.ScrapeGiftCodes(ctx)
				if err != nil {
					b.GetLogger().WithError(err).Error("Error during periodic scraping")
				} else {
					b.GetLogger().WithField("results", results).Info("Periodic scraping completed")
					// Notify only if there are new codes
					b.notifyIfNewCodes(results)
				}
				cancel()
			case <-b.ctx.Done():
				b.GetLogger().Info("Stopping periodic scraping")
				return
			}
		}
	}()
}

func (b *Bot) ScrapeGiftCodes(ctx context.Context) ([]ScrapeResult, error) {
	var results []ScrapeResult

	// Channels to collect results and errors from goroutines
	resultChan := make(chan ScrapeResult)
	errorChan := make(chan error)

	// Start scraping sites concurrently
	for _, site := range b.Config.Scrape.Sites {
		go func(site ScrapeSite) {
			codes, err := b.scrapeSite(ctx, site)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- ScrapeResult{
				SiteName: site.Name,
				Codes:    codes,
				Error:    err,
			}
		}(site) // Pass site as an argument to avoid race conditions
	}

	// Collect the results from all goroutines
	for range b.Config.Scrape.Sites {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case err := <-errorChan:
			b.GetLogger().WithError(err).Error("Error scraping site")
		}
	}

	// Update the internal state with the new codes
	newCodes := b.findNewCodes(results)
	if len(newCodes) > 0 {
		b.lastCheckedCodes = append(b.lastCheckedCodes, newCodes...) // Keep track of the new codes
	}

	return results, nil
}

// New function to notify if new codes were found
func (b *Bot) notifyIfNewCodes(results []ScrapeResult) {
	newCodes := b.findNewCodes(results)
	if len(newCodes) > 0 {
		b.GetLogger().Info("New gift codes found, notifying...")
		if err := b.notifyNewCodes(context.Background(), newCodes); err != nil {
			b.GetLogger().WithError(err).Error("Error notifying new codes")
		}
	} else {
		b.GetLogger().Info("No new gift codes found. Skipping notification.")
	}
}

func (b *Bot) scrapeSite(ctx context.Context, site ScrapeSite) ([]GiftCode, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", site.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	var codes []GiftCode
	doc.Find(site.Selector).Each(func(i int, s *goquery.Selection) {
		code := strings.TrimSpace(s.Text())
		description := strings.TrimSpace(s.Parent().Text())
		description = strings.TrimPrefix(description, code)
		description = strings.TrimSpace(description)
		codes = append(codes, GiftCode{Code: code, Description: description, Source: site.Name})
	})

	return codes, nil
}

func (b *Bot) findNewCodes(results []ScrapeResult) []GiftCode {
	var newCodes []GiftCode
	for _, result := range results {
		for _, code := range result.Codes {
			if !b.codeExists(code) {
				newCodes = append(newCodes, code)
			}
		}
	}
	return newCodes
}

func (b *Bot) codeExists(code GiftCode) bool {
	for _, existingCode := range b.lastCheckedCodes {
		if existingCode.Code == code.Code {
			return true
		}
	}
	return false
}

func (b *Bot) notifyNewCodes(ctx context.Context, newCodes []GiftCode) error {
	// Validate the Notification Channel ID
	channelID, err := strconv.ParseUint(b.Config.Discord.NotificationChannelID, 10, 64)
	if err != nil {
		b.GetLogger().WithError(err).WithField("channel_id", b.Config.Discord.NotificationChannelID).
			Error("Invalid Notification Channel ID provided. Please ensure it is a valid Discord channel ID.")
		return fmt.Errorf("invalid notification channel ID: %w", err)
	}

	// Prepare the message for new codes
	message := "🎉 New gift codes found:\n\n"
	for _, code := range newCodes {
		message += fmt.Sprintf("**Code:** %s\n**Description:** %s\n**Source:** %s\n\n", code.Code, code.Description, code.Source)
	}

	// Send the message to the notification channel
	_, err = b.Session.ChannelMessageSend(strconv.FormatUint(channelID, 10), message)
	if err != nil {
		return fmt.Errorf("error sending new codes notification: %w", err)
	}

	// Log successful notification
	b.GetLogger().WithField("code_count", len(newCodes)).Info("New gift codes notification sent")
	return nil
}
