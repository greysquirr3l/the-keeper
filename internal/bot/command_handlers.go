// command_handlers.go
package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func RegisterCommandHandlers() {
	RegisterCommand("id", handleIDCommand)
	RegisterCommand("term", handleTermCommand)
	RegisterCommand("giftcode", handleGiftcodeCommand)
	RegisterCommand("scrape", handleScrapeCommand)
	RegisterCommand("help", handleHelpCommand)

	// Register subcommands
	RegisterSubcommand("id", "add", handleIDAdd)
	RegisterSubcommand("id", "edit", handleIDEdit)
	RegisterSubcommand("id", "remove", handleIDRemove)
	RegisterSubcommand("id", "list", handleIDList)

	RegisterSubcommand("term", "add", handleTermAdd)
	RegisterSubcommand("term", "edit", handleTermEdit)
	RegisterSubcommand("term", "remove", handleTermRemove)
	RegisterSubcommand("term", "list", handleTermList)

	RegisterSubcommand("giftcode", "redeem", handleGiftcodeRedeem)
	RegisterSubcommand("giftcode", "deploy", handleGiftcodeDeploy)
	RegisterSubcommand("giftcode", "validate", handleGiftcodeValidate)
	RegisterSubcommand("giftcode", "list", handleGiftcodeList)
}

func handleIDCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !id <add|edit|remove|list> [arguments]")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Unknown subcommand for id. Use !help id for more information.")
}

func handleIDAdd(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !id add <playerID>")
		return
	}
	// Implement ID add logic
	s.ChannelMessageSend(m.ChannelID, "ID add functionality not implemented yet.")
}

func handleIDEdit(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !id edit <newPlayerID>")
		return
	}
	// Implement ID edit logic
	s.ChannelMessageSend(m.ChannelID, "ID edit functionality not implemented yet.")
}

func handleIDRemove(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !id remove <playerID>")
		return
	}
	// Implement ID remove logic
	s.ChannelMessageSend(m.ChannelID, "ID remove functionality not implemented yet.")
}

func handleIDList(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Implement ID list logic
	s.ChannelMessageSend(m.ChannelID, "ID list functionality not implemented yet.")
}

func handleTermCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !term <add|edit|remove|list> [arguments]")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Unknown subcommand for term. Use !help term for more information.")
}

func handleTermAdd(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !term add <title> <description>")
		return
	}
	// Implement term add logic
	s.ChannelMessageSend(m.ChannelID, "Term add functionality not implemented yet.")
}

func handleTermEdit(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !term edit <title> <new description>")
		return
	}
	// Implement term edit logic
	s.ChannelMessageSend(m.ChannelID, "Term edit functionality not implemented yet.")
}

func handleTermRemove(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !term remove <title>")
		return
	}
	// Implement term remove logic
	s.ChannelMessageSend(m.ChannelID, "Term remove functionality not implemented yet.")
}

func handleTermList(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Implement term list logic
	s.ChannelMessageSend(m.ChannelID, "Term list functionality not implemented yet.")
}

func handleGiftcodeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !giftcode <redeem|deploy|validate|list> <GiftCode>")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Unknown subcommand for giftcode. Use !help giftcode for more information.")
}

func handleGiftcodeRedeem(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !giftcode redeem <GiftCode>")
		return
	}
	// Implement giftcode redeem logic
	s.ChannelMessageSend(m.ChannelID, "Giftcode redeem functionality not implemented yet.")
}

func handleGiftcodeDeploy(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !giftcode deploy <GiftCode>")
		return
	}
	// Implement giftcode deploy logic
	s.ChannelMessageSend(m.ChannelID, "Giftcode deploy functionality not implemented yet.")
}

func handleGiftcodeValidate(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !giftcode validate <GiftCode>")
		return
	}
	// Implement giftcode validate logic
	s.ChannelMessageSend(m.ChannelID, "Giftcode validate functionality not implemented yet.")
}

func handleGiftcodeList(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Implement giftcode list logic
	s.ChannelMessageSend(m.ChannelID, "Giftcode list functionality not implemented yet.")
}

func handleScrapeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Implement scrape command logic
	s.ChannelMessageSend(m.ChannelID, "Scrape command not implemented yet.")
}

func handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	config, err := LoadCommandConfig("commands.yaml")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading command configuration.")
		return
	}

	if len(args) == 0 {
		// General help message
		helpMsg := "Available commands:\n"
		for cmdName, cmd := range config.Commands {
			if !cmd.Hidden {
				helpMsg += fmt.Sprintf("**%s%s**: %s\n", config.Prefix, cmdName, cmd.Description)
			}
		}
		helpMsg += fmt.Sprintf("\nUse `%shelp <command>` for more information on a specific command.", config.Prefix)
		s.ChannelMessageSend(m.ChannelID, helpMsg)
	} else {
		// Help for specific command
		cmdName := strings.ToLower(args[0])
		cmd, exists := config.Commands[cmdName]
		if !exists || cmd.Hidden {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No help available for command: %s", cmdName))
			return
		}

		helpMsg := fmt.Sprintf("**%s%s**: %s\n", config.Prefix, cmdName, cmd.Description)
		helpMsg += fmt.Sprintf("Usage: `%s`\n", cmd.Usage)
		if cmd.Cooldown != "" {
			helpMsg += fmt.Sprintf("Cooldown: %s\n", cmd.Cooldown)
		}

		if len(cmd.Subcommands) > 0 {
			helpMsg += "\nSubcommands:\n"
			for subCmdName, subCmd := range cmd.Subcommands {
				if !subCmd.Hidden {
					helpMsg += fmt.Sprintf("  **%s**: %s\n", subCmdName, subCmd.Description)
					helpMsg += fmt.Sprintf("    Usage: `%s`\n", subCmd.Usage)
					if subCmd.Cooldown != "" {
						helpMsg += fmt.Sprintf("    Cooldown: %s\n", subCmd.Cooldown)
					}
				}
			}
		}

		s.ChannelMessageSend(m.ChannelID, helpMsg)
	}
}
