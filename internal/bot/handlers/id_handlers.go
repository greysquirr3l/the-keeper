// File: internal/bot/handlers/id_handlers.go

package handlers

import (
	"fmt"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

func init() {
	bot.RegisterHandlerLater("handleIDCommand", handleIDCommand)
	bot.RegisterHandlerLater("handleIDAddCommand", handleIDAddCommand)
	bot.RegisterHandlerLater("handleIDRemoveCommand", handleIDRemoveCommand)
	bot.RegisterHandlerLater("handleIDEditCommand", handleIDEditCommand)
	bot.RegisterHandlerLater("handleIDListCommand", handleIDListCommand)
}

func handleIDCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: id <add|remove|edit|list> [args]")
		return
	}

	subCmdName := bot.NormalizeInput(args[0])
	switch subCmdName {
	case "add":
		handleIDAddCommand(s, m, args[1:], cmd)
	case "remove":
		handleIDRemoveCommand(s, m, args[1:], cmd)
	case "edit":
		handleIDEditCommand(s, m, args[1:], cmd)
	case "list":
		handleIDListCommand(s, m, args[1:], cmd)
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown subcommand: %s", subCmdName))
	}
}

func handleIDAddCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: id add <PlayerID>")
		return
	}

	playerID := bot.NormalizeInput(args[0])
	err := bot.AddPlayer(m.Author.ID, playerID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to add Player ID: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Player ID '%s' added successfully!", playerID))
}

func handleIDEditCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: id edit <newPlayerID>")
		return
	}

	newPlayerID := bot.NormalizeInput(args[0])
	err := bot.EditPlayerID(m.Author.ID, newPlayerID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to edit Player ID: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Your player ID has been updated to %s.", newPlayerID))
}

func handleIDRemoveCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	err := bot.RemovePlayer(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to remove Player ID: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Player ID removed successfully!")
}

func handleIDListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	players, err := bot.ListPlayers()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to list players: %v", err))
		return
	}

	if len(players) == 0 {
		s.ChannelMessageSend(m.ChannelID, "⚠️ No player IDs have been registered.")
		return
	}

	var response strings.Builder
	response.WriteString("Player ID List:\n")
	for _, player := range players {
		user, err := s.User(player.DiscordID)
		username := "Unknown User"
		if err == nil {
			username = user.Username
		}
		response.WriteString(fmt.Sprintf("%s: %s\n", username, player.PlayerID))
	}

	if err := bot.SendMessage(s, m.ChannelID, response.String()); err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Failed to send player ID list")
	}
}
