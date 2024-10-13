// File: internal/bot/tests/scrape_test.go

package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"the-keeper/internal/bot"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func setupMockBot() (*bot.Bot, sqlmock.Sqlmock) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(fmt.Sprintf("an error '%s' was not expected when opening a stub database connection", err))
	}

	mockConfig := &bot.Config{
		Discord: bot.DiscordConfig{
			Token:                 "mock_token",
			NotificationChannelID: "1234567890",
		},
	}

	mockSession := &discordgo.Session{}
	return bot.NewBot(mockConfig, mockSession, db), mock
}

func TestScrapeGiftCodes(t *testing.T) {
	b, _ := setupMockBot()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<html><body><strong>TESTCODE</strong></body></html>")
	}))
	defer server.Close()

	b.Config.Scrape.Sites = []bot.ScrapeSite{{
		Name:     "MockSite",
		URL:      server.URL,
		Selector: "strong",
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := b.ScrapeGiftCodes(ctx)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "MockSite", results[0].SiteName)
	assert.Len(t, results[0].Codes, 1)
	assert.Equal(t, "TESTCODE", results[0].Codes[0].Code)
}
