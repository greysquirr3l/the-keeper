// File: internal/bot/handlers/member_greeting.go

package handlers

import (
	"fmt"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

var welcomeMessageTemplate = `Welcome to the server, %s! 🎉

I am ** The Keeper **, a bot.

Please enter your Whiteout Survival PlayerID
to take full advantage ofmy features.`

func init() {
	bot.RegisterHandlerLater("handleWelcomeCommand", handleWelcomeCommand)
}

// Add the new member greeting handler to the bot instance
func AddNewMemberGreetingHandler(botInstance *bot.Bot) {
	botInstance.Session.AddHandler(handleNewMemberGreeting)
}

// Handle new member joining the server
func handleNewMemberGreeting(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	botInstance := bot.GetBot()
	if botInstance.GetLogger() == nil {
		return // Logger is not initialized, cannot proceed.
	}

	// Welcome message
	welcomeMessage := fmt.Sprintf(welcomeMessageTemplate, m.User.Username)
	channel, err := s.UserChannelCreate(m.User.ID)
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Failed to create DM channel with the new member")
		return
	}

	if _, err := s.ChannelMessageSend(channel.ID, welcomeMessage); err != nil {
		botInstance.GetLogger().WithError(err).Error("Failed to send welcome message to new member")
		return
	}

	// Listen for PlayerID response from the new member in their DM channel
	s.AddHandler(func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		// Ensure we are only processing messages from the new member in the DM channel
		if msg.Author.ID != m.User.ID || msg.ChannelID != channel.ID {
			return // Ignore messages from other users or other channels
		}

		playerID := strings.TrimSpace(msg.Content) // Assume the user's response is their PlayerID
		if playerID == "" {
			bot.SendMessage(s, msg.ChannelID, "⚠️ Please provide a valid PlayerID.")
			return
		}

		// Validate playerID format (should be numeric and between 3 and 12 digits)
		if len(playerID) < 3 || len(playerID) > 12 || !isNumeric(playerID) {
			bot.SendMessage(s, msg.ChannelID, "𐄂 Invalid playerID. It should be a number between 3 and 12 digits.")
			return
		}

		// Attempt to associate the PlayerID with the user's Discord ID
		handleIDAddCommand(s, msg, []string{playerID}, &bot.Command{})
		// No need to check for an error if the function does not return one.
		if err != nil {
			// Log the error for debugging
			botInstance.GetLogger().WithError(err).Error("Failed to add PlayerID")

			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				bot.SendMessage(s, msg.ChannelID, "⚠️ Unable to add playerID, it seems I already have it. If I don't, please let an admin know.")
			} else {
				bot.SendMessage(s, msg.ChannelID, "⚠️ An error occurred while adding your PlayerID. Please try again later or contact an admin.")
			}
			return
		}

		// Only send this success message if no error occurred
		bot.SendMessage(s, msg.ChannelID, "✓ PlayerID successfully associated! You are now ready to participate.")
	})
}

// Handle admin welcome command
func handleWelcomeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	if botInstance.GetLogger() == nil {
		bot.SendMessage(s, m.ChannelID, "⚠️ Logger is not initialized. Cannot proceed with this operation.")
		return
	}

	if !botInstance.IsAdmin(s, m.GuildID, m.Author.ID) {
		bot.SendMessage(s, m.ChannelID, "𐄂 You do not have permission to use this command.")
		return
	}

	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, "⚠️ Please provide a Discord ID or @mention to send the welcome message.")
		return
	}

	target := strings.TrimSpace(args[0])
	var userID string

	if strings.HasPrefix(target, "<@") && strings.HasSuffix(target, ">") {
		// Extract the user ID by removing <@ or <@! from the beginning, and > from the end
		userID = strings.TrimPrefix(target, "<@")
		userID = strings.TrimPrefix(userID, "!")
		userID = strings.TrimSuffix(userID, ">")
	} else {
		// Assume the input is a direct Discord user ID
		userID = target
	}

	member, err := s.GuildMember(m.GuildID, userID)
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Failed to find the specified user")
		bot.SendMessage(s, m.ChannelID, "⚠️ Failed to find the specified user. Please check the Discord ID or mention.")
		return
	}

	// Welcome message
	welcomeMessage := fmt.Sprintf(welcomeMessageTemplate, member.User.Username)
	channel, err := s.UserChannelCreate(member.User.ID)
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Failed to create DM channel with the user")
		bot.SendMessage(s, m.ChannelID, "⚠️ Unable to send a direct message to the user.")
		return
	}

	if _, err := s.ChannelMessageSend(channel.ID, welcomeMessage); err != nil {
		botInstance.GetLogger().WithError(err).Error("Failed to send welcome message to the user")
		return
	}

	// Add a handler to listen for the PlayerID response in the DM channel
	s.AddHandler(func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		// Ensure we are only processing messages from the intended user in the DM channel
		if msg.Author.ID != member.User.ID || msg.ChannelID != channel.ID {
			return // Ignore messages from other users or channels
		}

		playerID := strings.TrimSpace(msg.Content) // Assume the user's response is their PlayerID
		if playerID == "" {
			bot.SendMessage(s, msg.ChannelID, "⚠️ Please provide a valid PlayerID.")
			return
		}

		// Validate playerID format (should be numeric and between 3 and 12 digits)
		if len(playerID) < 3 || len(playerID) > 12 || !isNumeric(playerID) {
			bot.SendMessage(s, msg.ChannelID, "𐄂 Invalid playerID. It should be a number between 3 and 12 digits.")
			return
		}

		handleIDAddCommand(s, msg, []string{playerID}, &bot.Command{})
		if err != nil {
			// Log the error for debugging
			botInstance.GetLogger().WithError(err).Error("Failed to add PlayerID")

			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				bot.SendMessage(s, msg.ChannelID, "⚠️ Unable to add playerID, it seems I already have it. If I don't, please let an admin know.")
			} else {
				bot.SendMessage(s, msg.ChannelID, "⚠️ An error occurred while adding your PlayerID. Please try again later or contact an admin.")
			}
			return
		}

		// Only send this success message if no error occurred
		bot.SendMessage(s, msg.ChannelID, "✓ PlayerID successfully associated! You are now ready to participate.")
	})

	bot.SendMessage(s, m.ChannelID, fmt.Sprintf("✓ Welcome message successfully sent to %s.", member.User.Username))
}

// Helper function to check if a string is numeric
func isNumeric(str string) bool {
	for _, r := range str {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
