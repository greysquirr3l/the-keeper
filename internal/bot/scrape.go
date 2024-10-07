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

func (b *Bot) StartPeriodicScraping() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Adjust the interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if err := b.ScrapeGiftCodes(ctx); err != nil {
					b.logger.WithError(err).Error("Error during periodic scraping")
				}
				cancel()
			case <-b.ctx.Done():
				b.logger.Info("Stopping periodic scraping")
				return
			}
		}
	}()
}

func (b *Bot) ScrapeGiftCodes(ctx context.Context) error {
	b.GetLogger().Info("Starting gift code scraping")

	codesVG247, err := b.scrapeVG247Codes(ctx)
	if err != nil {
		b.GetLogger().WithError(err).Error("Error scraping VG247 gift codes")
	}

	codesLootbar, err := b.scrapeLootbarCodes(ctx)
	if err != nil {
		b.GetLogger().WithError(err).Error("Error scraping Lootbar gift codes")
	}

	allCodes := append(codesVG247, codesLootbar...)
	newCodes := b.findNewCodes(allCodes)
	if len(newCodes) > 0 {
		if err := b.notifyNewCodes(ctx, newCodes); err != nil {
			b.GetLogger().WithError(err).Error("Error notifying new codes")
		}
	}

	b.lastCheckedCodes = allCodes
	b.GetLogger().WithField("code_count", len(allCodes)).Info("Gift code scraping completed")
	return nil
}

func (b *Bot) scrapeVG247Codes(ctx context.Context) ([]GiftCode, error) {
	b.GetLogger().Info("Scraping VG247 for gift codes")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.vg247.com/whiteout-survival-codes", nil)
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
	doc.Find("ul li strong").Each(func(i int, s *goquery.Selection) {
		code := strings.TrimSpace(s.Text())
		description := strings.TrimSpace(s.Parent().Text())
		description = strings.TrimPrefix(description, code)
		description = strings.TrimSpace(description)
		codes = append(codes, GiftCode{Code: code, Description: description, Source: "VG247"})
	})

	b.GetLogger().WithField("code_count", len(codes)).Info("VG247 gift codes scraped")
	return codes, nil
}

func (b *Bot) scrapeLootbarCodes(ctx context.Context) ([]GiftCode, error) {
	b.GetLogger().Info("Scraping Lootbar for gift codes")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://lootbar.gg/blog/en/whiteout-survival-newest-codes.html", nil)
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
	doc.Find(".code-block").Each(func(i int, s *goquery.Selection) {
		code := strings.TrimSpace(s.Find(".code-block__code").Text())
		description := strings.TrimSpace(s.Find(".code-block__description").Text())
		codes = append(codes, GiftCode{Code: code, Description: description, Source: "Lootbar"})
	})

	b.GetLogger().WithField("code_count", len(codes)).Info("Lootbar gift codes scraped")
	return codes, nil
}

func (b *Bot) findNewCodes(currentCodes []GiftCode) []GiftCode {
	var newCodes []GiftCode
	for _, current := range currentCodes {
		isNew := true
		for _, last := range b.lastCheckedCodes {
			if current.Code == last.Code {
				isNew = false
				break
			}
		}
		if isNew {
			newCodes = append(newCodes, current)
		}
	}
	return newCodes
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
