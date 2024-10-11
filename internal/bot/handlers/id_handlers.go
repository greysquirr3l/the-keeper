// File: internal/bot/handlers/id_handlers.go
package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
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
	botInstance := bot.GetBot()
	isAuthorized := botInstance.IsAdmin(s, m.GuildID, m.Author.ID)

	// Add this logging
	botInstance.GetLogger().WithFields(logrus.Fields{
		"user_id":       m.Author.ID,
		"guild_id":      m.GuildID,
		"is_authorized": isAuthorized,
		"args_length":   len(args),
		"args":          args,
	}).Info("ID Add command invoked")

	var playerID, discordID string

	if isAuthorized && len(args) == 2 {
		// Authorized user adding for another player
		discordID = args[0]
		playerID = args[1]
	} else if len(args) == 1 {
		// Regular user adding for themselves
		discordID = m.Author.ID
		playerID = args[0]
	} else {
		usage := "Usage: `!id add <playerID>`"
		if isAuthorized {
			usage = "Usage: `!id add <discordID> <playerID>`"
		}
		bot.SendMessage(s, m.ChannelID, usage)
		return
	}

	if !playerIDRegex.MatchString(playerID) {
		bot.SendMessage(s, m.ChannelID, "𐄂 Invalid playerID. It should be a number between 3 and 12 digits.")
		return
	}

	err := botInstance.DB.Create(&bot.Player{DiscordID: discordID, PlayerID: playerID}).Error
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Error adding playerID")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ Unable to add playerID, it seems I already have it.  If I don't please let an admin know: %v", err))
		return
	}

	if discordID == m.Author.ID {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("✓ PlayerID %s has been added for you.", playerID))
	} else {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("✓ PlayerID %s has been added for Discord ID %s.", playerID, discordID))
	}
}

func handleIDEditCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, "Usage: `!id edit <newPlayerID>`")
		return
	}
	newPlayerID := args[0]
	if !playerIDRegex.MatchString(newPlayerID) {
		bot.SendMessage(s, m.ChannelID, "𐄂 Invalid playerID. It should be a number between 3 and 12 digits.")
		return
	}

	botInstance := bot.GetBot()
	err := botInstance.DB.Model(&bot.Player{}).Where("discord_id = ?", m.Author.ID).Update("player_id", newPlayerID).Error
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Error editing playerID")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ Error editing playerID: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Your playerID has been updated to %s.", newPlayerID))
}

func handleIDRemoveCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	err := botInstance.DB.Where("discord_id = ?", m.Author.ID).Delete(&bot.Player{}).Error
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Error removing playerID")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ Error removing playerID: %v", err))
		return
	}
	bot.SendMessage(s, m.ChannelID, "✓ Your playerID association has been removed.")
}

func handleIDListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	var players []bot.Player
	err := botInstance.DB.Order("discord_id").Find(&players).Error
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Error listing players")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ Error listing players: %v", err))
		return
	}
	if len(players) == 0 {
		bot.SendMessage(s, m.ChannelID, "⚠️ No playerIDs have been registered.")
		return
	}

	var response strings.Builder
	response.WriteString("PlayerID List:\n")
	for _, player := range players {
		user, err := s.User(player.DiscordID)
		username := "Unknown User"
		if err == nil {
			username = user.Username
		}
		response.WriteString(fmt.Sprintf("%s: %s\n", username, player.PlayerID))
	}
	if err := bot.SendMessage(s, m.ChannelID, response.String()); err != nil {
		botInstance.GetLogger().WithError(err).Error("Failed to send playerID list")
	}
}
