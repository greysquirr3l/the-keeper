package handlers

import (
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

// RegisterHandlers registers all command handlers with the bot instance
func RegisterHandlers(b *bot.Bot) {
	b.RegisterHandler("handleCommandName", handleCommandName)
	// Register any other handlers here...
}

func handleCommandName(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	bot.SendMessage(s, m.ChannelID, "Command not implemented yet.")
}
