// File: internal/bot/handlers/help_handlers.go

package handlers

import (
	"fmt"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

func init() {
	bot.RegisterHandlerLater("handleHelpCommand", handleHelpCommand)
	bot.RegisterHandlerLater("handleDumpDatabaseCommand", handleDumpDatabaseCommand) // New command registration
}

func handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) == 0 {
		sendGeneralHelp(s, m.ChannelID)
	} else {
		sendCommandHelp(s, m.ChannelID, args[0])
	}
}

func sendGeneralHelp(s *discordgo.Session, channelID string) {
	var helpMessage strings.Builder
	helpMessage.WriteString("Available commands:\n")
	for _, cmd := range bot.CommandRegistry {
		if !cmd.Hidden {
			helpMessage.WriteString(fmt.Sprintf("!%s: %s\n", cmd.Name, cmd.Description))
		}
	}
	helpMessage.WriteString("\nUse !help <command> for more information on a specific command.")

	if err := bot.SendMessage(s, channelID, helpMessage.String()); err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Failed to send general help message")
	}
}

func sendCommandHelp(s *discordgo.Session, channelID string, commandName string) {
	cmd, exists := bot.CommandRegistry[commandName]
	if !exists || cmd.Hidden {
		if err := bot.SendMessage(s, channelID, "Unknown command."); err != nil {
			bot.GetBot().GetLogger().WithError(err).Error("Failed to send unknown command message")
		}
		return
	}

	var helpMessage strings.Builder
	helpMessage.WriteString(fmt.Sprintf("Help for !%s:\n", cmd.Name))
	helpMessage.WriteString(fmt.Sprintf("Description: %s\n", cmd.Description))
	helpMessage.WriteString(fmt.Sprintf("Usage: %s\n", cmd.Usage))
	if cmd.Cooldown != "" {
		helpMessage.WriteString(fmt.Sprintf("Cooldown: %s\n", cmd.Cooldown))
	}
	if len(cmd.Subcommands) > 0 {
		helpMessage.WriteString("Subcommands:\n")
		for _, subCmd := range cmd.Subcommands {
			if !subCmd.Hidden {
				helpMessage.WriteString(fmt.Sprintf("  %s: %s\n", subCmd.Name, subCmd.Description))
			}
		}
	}

	if err := bot.SendMessage(s, channelID, helpMessage.String()); err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Failed to send command help message")
	}
}

// Dump the entire database (hidden, authorized command)
func handleDumpDatabaseCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	if !botInstance.IsAdmin(s, m.GuildID, m.Author.ID) {
		s.ChannelMessageSend(m.ChannelID, "𐄂 You do not have permission to use this command.")
		return
	}

	// Dump Terms table
	terms, err := botInstance.ListTerms()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to list terms: %v", err))
		return
	}

	if len(terms) == 0 {
		s.ChannelMessageSend(m.ChannelID, "⚠️ No terms available in the database.")
	} else {
		var response strings.Builder
		response.WriteString("Terms:\n")
		response.WriteString("| Term | Description |\n")
		response.WriteString("|------|-------------|\n")
		for _, term := range terms {
			response.WriteString(fmt.Sprintf("| %s | %s |\n", term.Term, term.Description))
		}
		s.ChannelMessageSend(m.ChannelID, response.String())
	}

	// Dump Player IDs table
	players, err := botInstance.ListPlayers()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to list players: %v", err))
		return
	}

	if len(players) == 0 {
		s.ChannelMessageSend(m.ChannelID, "⚠️ No players available in the database.")
	} else {
		var response strings.Builder
		response.WriteString("Players:\n")
		response.WriteString("| Discord ID | Player ID |\n")
		response.WriteString("|------------|-----------|\n")
		for _, player := range players {
			response.WriteString(fmt.Sprintf("| %s | %s |\n", player.DiscordID, player.PlayerID))
		}
		s.ChannelMessageSend(m.ChannelID, response.String())
	}
}
