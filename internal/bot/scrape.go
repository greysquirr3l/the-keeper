// File: internal/bot/scrape.go

package bot

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ScrapeSite struct {
	Name     string
	URL      string
	Selector string
}

type ScrapeResult struct {
	SiteName string
	Codes    []GiftCode
	Error    error
}

type GiftCode struct {
	Code        string
	Description string
	Source      string
}

func (b *Bot) ScrapeGiftCodes(ctx context.Context) ([]ScrapeResult, error) {
	var results []ScrapeResult

	for _, site := range b.Config.Scrape.Sites {
		codes, err := b.scrapeSite(ctx, site)
		results = append(results, ScrapeResult{
			SiteName: site.Name,
			Codes:    codes,
			Error:    err,
		})
	}

	newCodes := b.findNewCodes(results)
	if len(newCodes) > 0 {
		if err := b.notifyNewCodes(ctx, newCodes); err != nil {
			b.GetLogger().WithError(err).Error("Error notifying new codes")
		}
	}

	return results, nil
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
	message := "🎉 New gift codes found:\n\n"
	for _, code := range newCodes {
		message += fmt.Sprintf("**Code:** %s\n**Description:** %s\n**Source:** %s\n\n", code.Code, code.Description, code.Source)
	}

	channelID := b.Config.Discord.NotificationChannelID
	_, err := b.Session.ChannelMessageSend(channelID, message)
	if err != nil {
		return fmt.Errorf("error sending new codes notification: %w", err)
	}
	b.GetLogger().WithField("code_count", len(newCodes)).Info("New gift codes notification sent")
	return nil
}

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
				}
				cancel()
			case <-b.ctx.Done():
				b.GetLogger().Info("Stopping periodic scraping")
				return
			}
		}
	}()
}
