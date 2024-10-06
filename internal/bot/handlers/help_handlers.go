// File: internal/bot/handlers/help_handlers.go

package handlers

import (
	"fmt"
	"strings"

	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

func init() {
	bot.RegisterHandler("handleHelpCommand", handleHelpCommand)
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
	bot.SendMessage(s, channelID, helpMessage.String())
}

func sendCommandHelp(s *discordgo.Session, channelID string, commandName string) {
	cmd, exists := bot.CommandRegistry[commandName]
	if !exists || cmd.Hidden {
		bot.SendMessage(s, channelID, "Unknown command.")
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

	bot.SendMessage(s, channelID, helpMessage.String())
}
