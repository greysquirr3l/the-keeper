// File: internal/bot/handlers/scrape_handlers.go

package handlers

import (
	"context"
	"fmt"
	"strings"
	"the-keeper/internal/bot"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	bot.RegisterHandlerLater("handleScrapeCommand", handleScrapeCommand)
}

func handleScrapeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		botInstance.GetLogger().WithField("user", m.Author.Username).Info("Manual scraping initiated by user")
		results, err := botInstance.ScrapeGiftCodes(ctx)
		if err != nil {
			bot.SendMessage(s, m.ChannelID, fmt.Sprintf("❌ Scraping failed: %s", err.Error()))
			return
		}

		response := formatScrapeResults(results)
		bot.SendMessage(s, m.ChannelID, response)
	}()
}

func formatScrapeResults(results []bot.ScrapeResult) string {
	var sb strings.Builder
	sb.WriteString("📊 Scraping Results:\n\n")

	totalCodes := 0
	for _, result := range results {
		sb.WriteString(fmt.Sprintf("🌐 %s:\n", result.SiteName))
		if result.Error != nil {
			sb.WriteString(fmt.Sprintf("  ❌ Error: %s\n", result.Error))
		} else {
			sb.WriteString(fmt.Sprintf("  ✅ Codes found: %d\n", len(result.Codes)))
			for _, code := range result.Codes {
				sb.WriteString(fmt.Sprintf("    - %s: %s\n", code.Code, code.Description))
			}
			totalCodes += len(result.Codes)
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("📈 Total codes found: %d\n", totalCodes))

	return sb.String()
}
