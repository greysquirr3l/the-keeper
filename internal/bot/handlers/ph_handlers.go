// File: internal/bot/ph_handler.go

package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func init() {
	RegisterHandler("placeholderHandler", PlaceholderHandler)
}

// PlaceholderHandler is used for commands that are not yet implemented
func PlaceholderHandler(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	response := fmt.Sprintf("The command '%s' is not implemented yet... stay tuned!", cmd.Name)
	SendMessage(s, m.ChannelID, response)
}
