// File: internal/bot/tests/dump_database_test.go

package tests

import (
	"testing"

	"the-keeper/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDumpDatabaseCommand(t *testing.T) {
	b, mock := setupMockBot()
	mock.ExpectQuery("SELECT \* FROM players").WillReturnRows(sqlmock.NewRows([]string{"discord_id", "player_id"}).AddRow("12345", "54321"))

	s := &discordgo.Session{}
	m := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{ID: "adminID"},
		},
	}

	botInstance := bot.GetBot()
	botInstance.Admins = []string{"adminID"}
	handleDumpDatabaseCommand(s, m, []string{}, &bot.Command{})

	assert.Contains(t, s.LastMessage.Content, "| Discord ID | Player ID |")
	assert.Contains(t, s.LastMessage.Content, "| 12345 | 54321 |")
}
