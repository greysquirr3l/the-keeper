// File: internal/bot/handlers/giftcode_handlers.go

package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
)

func init() {
	bot.RegisterHandlerLater("handleGiftCodeCommand", handleGiftCodeCommand)
	bot.RegisterHandlerLater("handleGiftCodeRedeemCommand", handleGiftCodeRedeemCommand)
	bot.RegisterHandlerLater("handleGiftCodeDeployCommand", handleGiftCodeDeployCommand)
	bot.RegisterHandlerLater("handleGiftCodeValidateCommand", handleGiftCodeValidateCommand)
	bot.RegisterHandlerLater("handleGiftCodeListCommand", handleGiftCodeListCommand)
}

// Send help message for gift code commands
func sendGiftCodeHelp(s *discordgo.Session, channelID string, cmd *bot.Command) {
	helpMessage := "Available giftcode subcommands:\n"
	for name, subCmd := range cmd.Subcommands {
		if !subCmd.Hidden {
			helpMessage += fmt.Sprintf("  ** %s** : %s\n", name, subCmd.Description)
			helpMessage += fmt.Sprintf("    Usage: `%s`\n", subCmd.Usage)
		}
	}
	if err := bot.SendMessage(s, channelID, helpMessage); err != nil {
		bot.GetBot().GetLogger().WithError(err).Error("Failed to send gift code help message")
	}
}

// Main handler for gift code commands
func handleGiftCodeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	if botInstance.GetLogger() == nil {
		bot.SendMessage(s, m.ChannelID, "⚠️ Logger is not initialized. Cannot proceed with this operation.")
		return
	}

	if len(args) == 0 {
		sendGiftCodeHelp(s, m.ChannelID, cmd)
		return
	}

	subCmdName := bot.NormalizeInput(args[0])
	subCmd, exists := cmd.Subcommands[subCmdName]
	if !exists {
		bot.SendMessage(s, m.ChannelID, "⚠️ Unknown subcommand. Use !help giftcode to see available subcommands.")
		return
	}

	if subCmd.HandlerFunc != nil {
		subCmd.HandlerFunc(s, m, args[1:], subCmd)
	} else {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ The subcommand '%s' is not implemented yet.", subCmdName))
	}
}

// Deploy gift code command handler
func handleGiftCodeDeployCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	logger := botInstance.GetLogger()
	if logger == nil {
		bot.SendMessage(s, m.ChannelID, "⚠️ Logger is not initialized. Cannot proceed with this operation.")
		return
	}

	if !botInstance.IsAdmin(s, m.GuildID, m.Author.ID) {
		bot.SendMessage(s, m.ChannelID, "𐄂 You do not have permission to use this command.")
		return
	}

	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Usage: `%s`", cmd.Usage))
		return
	}

	giftCode := strings.TrimSpace(args[0]) // Keep the original case, only trim spaces.
	playerIDs, err := botInstance.GetAllPlayerIDs()
	if err != nil {
		logger.WithError(err).Error("Error retrieving Player IDs")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("𐄂 Error retrieving Player IDs: %v", err))
		return
	}

	if len(playerIDs) == 0 {
		bot.SendMessage(s, m.ChannelID, "⚠️ No player IDs available for deployment.")
		return
	}

	bot.SendMessage(s, m.ChannelID, "\n\n🚀 Deploying gift code to all users...")

	// Deploy the gift code using concurrent processing
	go func() {
		for discordID, playerID := range playerIDs {
			success, message, err := botInstance.RedeemGiftCode(playerID, giftCode)
			if err != nil {
				logger.WithError(err).
					WithField("player_id", playerID).
					WithField("gift_code", giftCode).
					Error("Error redeeming gift code")
				bot.SendMessage(s, m.ChannelID, fmt.Sprintf("𐄂 Error for Player ID %s: %v", playerID, err))
				continue
			}

			status := "Success"
			if !success {
				status = "Failed"
			}

			// Handle specific error codes from the API
			switch message {
			case "Gift Code not found":
				status = "Failed"
			case "Expired, unable to claim":
				status = "Failed"
			case "Gift code already claimed":
				status = "Already Claimed"
			case "API request failed: 40014":
				message = "Gift Code not found"
				status = "Failed"
			case "API request failed: 40007":
				message = "Expired, unable to claim"
				status = "Failed"
			case "API request failed: 40008":
				message = "Gift code already claimed"
				status = "Already Claimed"
			}

			err = botInstance.RecordGiftCodeRedemption(discordID, playerID, giftCode, status)
			if err != nil {
				logger.WithError(err).Error("Gift code redeemed but failed to record in database")
				bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ Gift code redeemed for Player ID %s but failed to record: %v", playerID, err))
				continue
			}

			bot.SendMessage(s, m.ChannelID, fmt.Sprintf("-# Player ID ** %s** : %s", playerID, message))
		}
		bot.SendMessage(s, m.ChannelID, "\n✓ Gift code deployment completed.")
	}()
}

// Redeem gift code command handler
func handleGiftCodeRedeemCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	if botInstance.GetLogger() == nil {
		bot.SendMessage(s, m.ChannelID, "⚠️ Logger is not initialized. Cannot proceed with this operation.")
		return
	}

	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Usage: %s", cmd.Usage))
		return
	}

	giftCode := strings.TrimSpace(args[0]) // Keep the original case, only trim spaces.
	var playerID string
	var discordID string = m.Author.ID

	// Check if an extra argument (discordID) is provided for authorized users
	if len(args) > 1 && botInstance.IsAdmin(s, m.GuildID, m.Author.ID) {
		discordID = args[1]
		playerID, _ = botInstance.GetPlayerID(discordID)
		if playerID == "" {
			bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ The provided Discord ID '%s' does not have an associated Player ID.", discordID))
			return
		}
	} else {
		var err error
		playerID, err = botInstance.GetPlayerID(m.Author.ID)
		if err != nil {
			bot.SendMessage(s, m.ChannelID, "⚠️ You do not have a Player ID associated. Use `!id add <PlayerID>` to associate your account.")
			return
		}
	}

	success, message, err := botInstance.RedeemGiftCode(playerID, giftCode)
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Error redeeming gift code")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("𐄂 Error redeeming gift code: %v", err))
		return
	}

	status := "Success"
	if !success {
		status = "Failed"
	}

	// Handle specific error codes from the API
	switch message {
	case "Gift Code not found":
		status = "Failed"
	case "Expired, unable to claim":
		status = "Failed"
	case "Gift code already claimed":
		status = "Already Claimed"
	case "API request failed: 40014":
		message = "Gift Code not found"
		status = "Failed"
	case "API request failed: 40007":
		message = "Expired, unable to claim"
		status = "Failed"
	case "API request failed: 40008":
		message = "Gift code already claimed"
		status = "Already Claimed"
	}

	err = botInstance.RecordGiftCodeRedemption(discordID, playerID, giftCode, status)
	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Gift code redeemed but failed to record")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("⚠️ Gift code redeemed but failed to record: %v", err))
		return
	}

	bot.SendMessage(s, m.ChannelID, message)
}

// Validate gift code command handler
func handleGiftCodeValidateCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	if botInstance.GetLogger() == nil {
		bot.SendMessage(s, m.ChannelID, "⚠️ Logger is not initialized. Cannot proceed with this operation.")
		return
	}

	if len(args) < 1 {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("Usage: `%s`", cmd.Usage))
		return
	}

	giftCode := strings.TrimSpace(args[0]) // Keep the original case, only trim spaces.
	playerID, err := botInstance.GetPlayerID(m.Author.ID)
	if err != nil {
		bot.SendMessage(s, m.ChannelID, "𐄂 You do not have a Player ID associated. Use `!id add <PlayerID>` to associate your account.")
		return
	}

	isValid, message := botInstance.ValidateGiftCode(giftCode, playerID)
	if isValid {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("✓ Gift code `%s` is valid.", giftCode))
	} else {
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("𐄂 Invalid gift code: %s", message))
	}
}

// List all gift code redemptions command handler
func handleGiftCodeListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *bot.Command) {
	botInstance := bot.GetBot()
	if botInstance.GetLogger() == nil {
		bot.SendMessage(s, m.ChannelID, "⚠️ Logger is not initialized. Cannot proceed with this operation.")
		return
	}

	page := 1
	itemsPerPage := 10

	if len(args) > 0 {
		if p, err := strconv.Atoi(args[0]); err == nil {
			page = p
		}
	}

	isAdmin := botInstance.IsAdmin(s, m.GuildID, m.Author.ID)

	var redemptions []bot.GiftCodeRedemption
	var err error

	if isAdmin {
		redemptions, err = botInstance.GetAllGiftCodeRedemptionsPaginated(page, itemsPerPage)
	} else {
		redemptions, err = botInstance.GetUserGiftCodeRedemptionsPaginated(m.Author.ID, page, itemsPerPage)
	}

	if err != nil {
		botInstance.GetLogger().WithError(err).Error("Error retrieving gift codes")
		bot.SendMessage(s, m.ChannelID, fmt.Sprintf("𐄂 Error retrieving gift codes: %v", err))
		return
	}

	if len(redemptions) == 0 {
		bot.SendMessage(s, m.ChannelID, "⚠️ No gift codes found for this page.")
		return
	}

	message := fmt.Sprintf("📜 Gift code redemptions (Page %d):\n", page)
	for _, r := range redemptions {
		if isAdmin {
			message += fmt.Sprintf("Discord ID: %s, Player ID: %s, Code: %s, Status: %s\n", r.DiscordID, r.PlayerID, r.GiftCode, r.Status)
		} else {
			message += fmt.Sprintf("Code: %s, Status: %s\n", r.GiftCode, r.Status)
		}
	}
	message += fmt.Sprintf("\nUse '!giftcode list %d' to see the next page", page+1)

	bot.SendMessage(s, m.ChannelID, message)
}
