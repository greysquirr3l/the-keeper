// File: internal/bot/handlers/id_handlers.go
package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

var (
	playerIDRegex = regexp.MustCompile(`^\d{3,12}$`)
)

func init() {
	bot.RegisterHandlerLater("handleIDCommand", handleIDCommand)
	bot.RegisterHandlerLater("handleIDAddCommand", handleIDAddCommand)
	bot.RegisterHandlerLater("handleIDEditCommand", handleIDEditCommand)
	bot.RegisterHandlerLater("handleIDRemoveCommand", handleIDRemoveCommand)
	bot.RegisterHandlerLater("handleIDListCommand", handleIDListCommand)
}

func handleIDCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) == 0 {
		sendIDHelp(s, m.ChannelID, cmd)
		return
	}

	subCmdName := args[0]
	subCmd, exists := cmd.Subcommands[subCmdName]
	if !exists {
		bot.SendMessage(s, m.ChannelID, "Unknown subcommand. Use `!help id` to see available subcommands.")
		return
	}

	if subCmd.HandlerFunc != nil {
		subCmd.HandlerFunc(s, m, args[1:], subCmd)
	} else {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("The subcommand `%s` is not implemented yet.", subCmdName))
	}
}

func sendIDHelp(s *discordgo.Session, channelID string, cmd *bot.Command) {
	var helpMsg strings.Builder
	helpMsg.WriteString("Available ID subcommands:\n")
	for subName, subCmd := range cmd.Subcommands {
		if !subCmd.Hidden {
			helpMsg.WriteString(fmt.Sprintf("  %s: %s\n", subName, subCmd.Description))
			helpMsg.WriteString(fmt.Sprintf("    Usage: %s\n", subCmd.Usage))
			if subCmd.Cooldown != "" {
				helpMsg.WriteString(fmt.Sprintf("    Cooldown: %s\n", subCmd.Cooldown))
			}
			helpMsg.WriteString("\n")
		}
	}
	if err := bot.SendMessage(s, channelID, helpMsg.String()); err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Failed to send ID help message")
	}
}

func handleIDAddCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, "Usage: !id add <playerID>")
		return
	}
	playerID := args[0]
	if !playerIDRegex.MatchString(playerID) {
		bot.SendMessage(s, m.ChannelID, "Invalid playerID. It should be a number between 3 and 12 digits.")
		return
	}
	err := bot.AddPlayer(m.Author.ID, playerID)
	if err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Error adding player ID")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error adding player ID: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Player ID %s has been added for user %s.", playerID, m.Author.Username))
}

func handleIDEditCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, "Usage: !id edit <newPlayerID>")
		return
	}
	newPlayerID := args[0]
	if !playerIDRegex.MatchString(newPlayerID) {
		bot.SendMessage(s, m.ChannelID, "Invalid playerID. It should be a number between 3 and 12 digits.")
		return
	}
	err := bot.EditPlayerID(m.Author.ID, newPlayerID)
	if err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Error editing player ID")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error editing player ID: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Your player ID has been updated to %s.", newPlayerID))
}

func handleIDRemoveCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	err := bot.RemovePlayer(m.Author.ID)
	if err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Error removing player ID")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error removing player ID: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, "Your player ID association has been removed.")
}

func handleIDListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	players, err := bot.ListPlayers()
	if err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Error listing players")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error listing players: %v", err))
		return
	}
	if len(players) == 0 {
		bot.SendMessage(s, m.ChannelID, "No player IDs have been registered.")
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
