package bot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Bot structure
type Bot struct {
	Config       *Config
	Session      *discordgo.Session
	Logger       *logrus.Logger
	shutdownChan chan struct{}
}

// NewBot creates a new bot instance
func NewBot(config *Config) (*Bot, error) {
	logger := logrus.New()

	bot := &Bot{
		Config:       config,
		Logger:       logger,
		shutdownChan: make(chan struct{}),
	}

	if config.Discord.Enabled {
		session, err := discordgo.New("Bot " + config.Discord.Token)
		if err != nil {
			return nil, fmt.Errorf("error creating Discord session: %w", err)
		}
		bot.Session = session
	}

	return bot, nil
}

// Start launches the Discord bot (if enabled)
func (b *Bot) Start(ctx context.Context) error {
	if b.Config.Discord.Enabled {
		b.Logger.Info("Starting Discord bot")
		err := b.initDiscord(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize Discord: %w", err)
		}
	}
	return nil
}

// Shutdown gracefully shuts down the bot
func (b *Bot) Shutdown() error {
	if b.Config.Discord.Enabled {
		if err := b.Session.Close(); err != nil {
			b.Logger.WithError(err).Error("Error closing Discord session")
		}
	}
	b.Logger.Info("Bot has been shut down")
	return nil
}

// HealthCheckHandler handles /healthz endpoint
func (b *Bot) HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}
}

// HandleOAuth2Callback handles the OAuth2 callback for Discord authorization
func (b *Bot) HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OAuth2 callback received!")
}

// initDiscord sets up the Discord session handlers
func (b *Bot) initDiscord(ctx context.Context) error {
	b.Session.AddHandler(b.onReady)

	// Open the session
	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %w", err)
	}

	b.Logger.Info("Discord bot is running")
	return nil
}

// onReady handler for when the bot is ready
func (b *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	b.Logger.Info("Bot is now connected to Discord")
}
