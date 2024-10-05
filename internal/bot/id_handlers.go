// File: internal/bot/id_handlers.go

package bot

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	playerIDRegex = regexp.MustCompile(`^\d{3,12}$`)
	playerIDs     = make(map[string]string) // map[DiscordID]PlayerID
	playerIDMutex sync.RWMutex
)

func handleIDAddCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	if len(args) < 1 {
		SendMessage(s, m.ChannelID, "Usage: !id add <playerID>")
		return
	}

	playerID := args[0]
	if !playerIDRegex.MatchString(playerID) {
		SendMessage(s, m.ChannelID, "Invalid playerID. It should be a number between 3 and 12 digits.")
		return
	}

	playerIDMutex.Lock()
	playerIDs[m.Author.ID] = playerID
	playerIDMutex.Unlock()

	SendMessage(s, m.ChannelID, fmt.Sprintf("Player ID %s has been added for user %s.", playerID, m.Author.Username))
}

func handleIDEditCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	if len(args) < 1 {
		SendMessage(s, m.ChannelID, "Usage: !id edit <newPlayerID>")
		return
	}

	newPlayerID := args[0]
	if !playerIDRegex.MatchString(newPlayerID) {
		SendMessage(s, m.ChannelID, "Invalid playerID. It should be a number between 3 and 12 digits.")
		return
	}

	playerIDMutex.Lock()
	playerIDs[m.Author.ID] = newPlayerID
	playerIDMutex.Unlock()

	SendMessage(s, m.ChannelID, fmt.Sprintf("Your player ID has been updated to %s.", newPlayerID))
}

func handleIDRemoveCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	playerIDMutex.Lock()
	delete(playerIDs, m.Author.ID)
	playerIDMutex.Unlock()

	SendMessage(s, m.ChannelID, "Your player ID has been removed.")
}

func handleIDListCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	playerIDMutex.RLock()
	defer playerIDMutex.RUnlock()

	if len(playerIDs) == 0 {
		SendMessage(s, m.ChannelID, "No player IDs have been registered.")
		return
	}

	var response strings.Builder
	response.WriteString("Player ID List:\n")
	for discordID, playerID := range playerIDs {
		user, err := s.User(discordID)
		username := "Unknown User"
		if err == nil {
			username = user.Username
		}
		response.WriteString(fmt.Sprintf("%s: %s\n", username, playerID))
	}

	SendMessage(s, m.ChannelID, response.String())
}
