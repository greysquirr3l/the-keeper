// File: internal/bot/handlers/term_handlers.go

package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

var termRegex = regexp.MustCompile(`^\S+$`)

func init() {
	bot.RegisterHandlerLater("handleTermCommand", handleTermCommand)
	bot.RegisterHandlerLater("handleTermAddCommand", handleTermAddCommand)
	bot.RegisterHandlerLater("handleTermEditCommand", handleTermEditCommand)
	bot.RegisterHandlerLater("handleTermRemoveCommand", handleTermRemoveCommand)
	bot.RegisterHandlerLater("handleTermListCommand", handleTermListCommand)
	bot.RegisterHandlerLater("handleTermGetCommand", handleTermGetCommand)
}

// Main term command handler
func handleTermCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!term <add|edit|remove|list|get> [args]`")
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
		handleTermGetCommand(s, m, args, cmd)
	}
}

// Add a new term
func handleTermAddCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!term add <term> <description>`")
		return
	}

	term := args[0]
	if !termRegex.MatchString(term) {
		s.ChannelMessageSend(m.ChannelID, "𐄂 Invalid term. The term must be a single word without spaces.")
		return
	}

	// Replace literal `\n` with actual newlines to preserve formatting
	description := strings.Join(args[1:], " ")
	description = strings.ReplaceAll(description, `\n`, "\n")

	// Add the term with preserved markdown and newlines
	err := bot.GetBot().AddTerm(term, description)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to add term: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✓ Term '%s' added successfully!", term))
}

// Edit an existing term
func handleTermEditCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!term edit <term> <new description>`")
		return
	}

	term := args[0]
	if !termRegex.MatchString(term) {
		s.ChannelMessageSend(m.ChannelID, "𐄂 Invalid term. The term must be a single word without spaces.")
		return
	}

	// Join the arguments to form the new description, preserving the input formatting
	newDescription := strings.Join(args[1:], " ")

	// Replace literal `\n` strings with actual newline characters
	newDescription = strings.ReplaceAll(newDescription, `\n`, "\n")

	// Debugging log to check the final formatted description before saving
	log.Printf("Editing term '%s' with description: %s", term, newDescription)

	// Edit the term in the database
	err := bot.GetBot().EditTerm(term, newDescription)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to edit term: %v", err))
		return
	}

	// Send a success message including the updated description with proper markdown
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✓ Term '%s' updated successfully!\n%s", term, newDescription))
}

// Delete an existing term
func handleTermRemoveCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!term remove <term>`")
		return
	}

	term := args[0]
	if !termRegex.MatchString(term) {
		s.ChannelMessageSend(m.ChannelID, "𐄂 Invalid term. The term must be a single word without spaces.")
		return
	}
	err := bot.GetBot().RemoveTerm(term)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to remove term: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✓ Term '%s' removed successfully!", term))
}

// List all terms
func handleTermListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	terms, err := bot.GetBot().ListTerms()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to list terms: %v", err))
		return
	}

	if len(terms) == 0 {
		s.ChannelMessageSend(m.ChannelID, "⚠️ No terms available.")
		return
	}

	var response strings.Builder
	response.WriteString("Terms:\n")
	for _, term := range terms {
		response.WriteString(fmt.Sprintf("- %s\n", term.Term))
	}

	s.ChannelMessageSend(m.ChannelID, response.String())
}

// Get a term's description
func handleTermGetCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: `!term get <term>`")
		return
	}

	term := args[0]
	if !termRegex.MatchString(term) {
		s.ChannelMessageSend(m.ChannelID, "𐄂 Invalid term. The term must be a single word without spaces.")
		return
	}
	description, err := bot.GetBot().GetTermDescription(term)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Failed to get term description: %v", err))
		return
	}

	// s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```%s```", description))
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s", description))
}
