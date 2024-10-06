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

func handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	if len(args) == 0 {
		// General help
		var helpMessage strings.Builder
		helpMessage.WriteString("Available commands:\n")

		for name, cmd := range CommandRegistry {
			if !cmd.Hidden {
				helpMessage.WriteString(fmt.Sprintf("%s: %s\n", name, cmd.Description))
			}
		}

		helpMessage.WriteString("\nUse !help <command> for more information on a specific command.")
		SendMessage(s, m.ChannelID, helpMessage.String())
	} else {
		// Specific command help
		cmdName := args[0]
		if cmd, exists := CommandRegistry[cmdName]; exists && !cmd.Hidden {
			var helpMessage strings.Builder
			helpMessage.WriteString(fmt.Sprintf("Help for %s:\n", cmdName))
			helpMessage.WriteString(fmt.Sprintf("Description: %s\n", cmd.Description))
			helpMessage.WriteString(fmt.Sprintf("Usage: %s\n", cmd.Usage))
			if cmd.Cooldown != "" {
				helpMessage.WriteString(fmt.Sprintf("Cooldown: %s\n", cmd.Cooldown))
			}
			if len(cmd.Subcommands) > 0 {
				helpMessage.WriteString("Subcommands:\n")
				for subName, subCmd := range cmd.Subcommands {
					if !subCmd.Hidden {
						helpMessage.WriteString(fmt.Sprintf("  %s: %s\n", subName, subCmd.Description))
					}
				}
			}
			SendMessage(s, m.ChannelID, helpMessage.String())
		} else {
			SendMessage(s, m.ChannelID, "Unknown command. Use !help to see available commands.")
		}
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
