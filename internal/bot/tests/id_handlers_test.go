// File: internal/bot/tests/id_handlers_test.go

package tests

import (
	"testing"

	"the-keeper/internal/bot"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestHandleIDAddCommand(t *testing.T) {
	b, mock := setupMockBot()
	mock.ExpectExec("INSERT INTO players").WithArgs("12345", "54321").WillReturnResult(sqlmock.NewResult(1, 1))

	s := &discordgo.Session{}
	m := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author:  &discordgo.User{ID: "12345"},
			Content: "!id add 54321",
		},
	}

	err := bot.HandleIDAddCommand(s, m, []string{"54321"}, &bot.Command{})
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
