// File: internal/bot/handlers/ph_handler.go

package handlers

import (
	"fmt"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

func init() {
	bot.RegisterHandlerLater("placeholderHandler", PlaceholderHandler)
}

// PlaceholderHandler is used for commands that are not yet implemented
func PlaceholderHandler(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	response := fmt.Sprintf("The command '%s' is not implemented yet... stay tuned!", cmd.Name)
	if err := bot.SendMessage(s, m.ChannelID, response); err != nil {
		bot.GetBot().Logger.WithError(err).Error("Failed to send placeholder message")
	}
}
