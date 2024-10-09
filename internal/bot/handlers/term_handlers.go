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

// Main term command handler
func handleTermCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: term <add|edit|remove|list> [args]")
		return
	}

	subcommand := bot.NormalizeInput(args[0])
	switch subcommand {
	case "add":
		handleTermAddCommand(s, m, args[1:], cmd)
	case "edit":
		handleTermEditCommand(s, m, args[1:], cmd)
	case "remove":
		handleTermRemoveCommand(s, m, args[1:], cmd)
	case "list":
		handleTermListCommand(s, m, args[1:], cmd)
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown subcommand: %s", subcommand))
	}
}

// Add a new term
func handleTermAddCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: term add <term> <definition>")
		return
	}

	term := bot.NormalizeInput(args[0])
	definition := bot.NormalizeInput(strings.Join(args[1:], " "))

	err := bot.AddTerm(term, definition)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to add term: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Term '%s' added successfully!", term))
}

// Edit an existing term
func handleTermEditCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: term edit <term> <new definition>")
		return
	}

	term := bot.NormalizeInput(args[0])
	newDefinition := strings.Join(args[1:], " ")

	err := bot.EditTerm(term, newDefinition)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to edit term: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Term '%s' updated successfully!", term))
}

// Delete an existing term
func handleTermRemoveCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: term remove <term>")
		return
	}

	term := bot.NormalizeInput(args[0])

	err := bot.RemoveTerm(term)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to remove term: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Term '%s' removed successfully!", term))
}

// List all terms
func handleTermListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	terms, err := bot.ListTerms()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to list terms: %v", err))
		return
	}

	if len(terms) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No terms available.")
		return
	}

	var response strings.Builder
	response.WriteString("Terms:\n")
	for _, term := range terms {
		response.WriteString(fmt.Sprintf("- %s: %s\n", term.Term, term.Description))
	}

	s.ChannelMessageSend(m.ChannelID, response.String())
}
