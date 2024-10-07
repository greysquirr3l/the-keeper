// File: ./internal/bot/handlers/term_handlers.go

package handlers

import (
	"fmt"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

func init() {
	bot.RegisterHandlerLater("handleTermCommand", handleTermCommand)
	bot.RegisterHandlerLater("handleTermAddCommand", handleTermAddCommand)
	bot.RegisterHandlerLater("handleTermEditCommand", handleTermEditCommand)
	bot.RegisterHandlerLater("handleTermRemoveCommand", handleTermRemoveCommand)
	bot.RegisterHandlerLater("handleTermListCommand", handleTermListCommand)
}

func handleTermCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) == 0 {
		sendTermHelp(s, m.ChannelID, cmd)
		return
	}
	subCmd, exists := cmd.Subcommands[args[0]]
	if !exists {
		bot.SendMessage(s, m.ChannelID, "Unknown subcommand. Use !help term to see available subcommands.")
		return
	}
	if subCmd.HandlerFunc != nil {
		subCmd.HandlerFunc(s, m, args[1:], subCmd)
	} else {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("The subcommand '%s' is not implemented yet.", args[0]))
	}
}

func sendTermHelp(s *discordgo.Session, channelID string, cmd *bot.Command) {
	helpMessage := "Available term subcommands:\n"
	for name, subCmd := range cmd.Subcommands {
		if !subCmd.Hidden {
			helpMessage += fmt.Sprintf("  %s: %s\n", name, subCmd.Description)
			helpMessage += fmt.Sprintf("    Usage: %s\n", subCmd.Usage)
		}
	}
	if err := bot.SendMessage(s, channelID, helpMessage); err != nil {
		bot.GetBot().Logger.WithError(err).Error("Failed to send term help message")
	}
}

func handleTermAddCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if !bot.IsAuthorized(s, m.GuildID, m.Author.ID) {
		bot.SendMessage(s, m.ChannelID, "You don't have permission to use this command.")
		return
	}
	if len(args) < 2 {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Usage: %s", cmd.Usage))
		return
	}
	term := args[0]
	description := strings.Join(args[1:], " ")
	err := bot.AddTerm(term, description)
	if err != nil {
		bot.GetBot().Logger.WithError(err).Error("Failed to add term")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error adding term: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Term '%s' has been added.", term))
}

func handleTermEditCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if !bot.IsAuthorized(s, m.GuildID, m.Author.ID) {
		bot.SendMessage(s, m.ChannelID, "You don't have permission to use this command.")
		return
	}
	if len(args) < 2 {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Usage: %s", cmd.Usage))
		return
	}
	term := args[0]
	newDescription := strings.Join(args[1:], " ")
	err := bot.EditTerm(term, newDescription)
	if err != nil {
		bot.GetBot().Logger.WithError(err).Error("Failed to edit term")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error editing term: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Term '%s' has been updated.", term))
}

func handleTermRemoveCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if !bot.IsAuthorized(s, m.GuildID, m.Author.ID) {
		bot.SendMessage(s, m.ChannelID, "You don't have permission to use this command.")
		return
	}
	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Usage: %s", cmd.Usage))
		return
	}
	term := args[0]
	err := bot.RemoveTerm(term)
	if err != nil {
		bot.GetBot().Logger.WithError(err).Error("Failed to remove term")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error removing term: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Term '%s' has been removed.", term))
}

func handleTermListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	terms, err := bot.ListTerms()
	if err != nil {
		bot.GetBot().Logger.WithError(err).Error("Failed to list terms")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Error listing terms: %v", err))
		return
	}
	if len(terms) == 0 {
		bot.SendMessage(s, m.ChannelID, "No terms have been added yet.")
		return
	}
	var response strings.Builder
	response.WriteString("Term List:\n")
	for _, term := range terms {
		response.WriteString(fmt.Sprintf("%s: %s\n", term.Term, term.Description))
	}
	if err := bot.SendMessage(s, m.ChannelID, response.String()); err != nil {
		bot.GetBot().Logger.WithError(err).Error("Failed to send term list")
	}
}
