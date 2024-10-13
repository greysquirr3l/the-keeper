// File: internal/bot/tests/member_greeting_test.go

package tests

import (
	"testing"

	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestWelcomeMessage(t *testing.T) {
	b, _ := setupMockBot()
	mockSession := &discordgo.Session{}
	newMember := &discordgo.GuildMemberAdd{
		User: &discordgo.User{
			ID:       "12345",
			Username: "NewUser",
		},
	}

	bot.AddNewMemberGreetingHandler(b)
	go handleNewMemberGreeting(mockSession, newMember)

	// Test if welcome message is sent correctly
	assert.Equal(t, "Welcome to the server, NewUser! 🎉\nPlease enter your Whiteout Survival PlayerID to take full advantage of\nmy features.", mockSession.LastMessage.Content)
}
