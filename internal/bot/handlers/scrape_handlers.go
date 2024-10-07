// File: internal/bot/handlers/scrape_handlers.go

package handlers

import (
	"context"
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
		// Create a new context with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		botInstance.GetLogger().WithField("user", m.Author.Username).Info("Manual scraping initiated by user")
		err := botInstance.ScrapeGiftCodes(ctx)
		if err != nil {
			botInstance.SendMessage(s, m.ChannelID, "❌ Scraping failed: "+err.Error())
			return
		}
		botInstance.SendMessage(s, m.ChannelID, "✅ Scraping completed successfully.")
	}()
}
