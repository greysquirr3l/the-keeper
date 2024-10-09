// File: internal/bot/bot.go

package bot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// CommandHandler is the function type used to handle bot commands
type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate, []string, *Command)

var instance *Bot
var pendingHandlers = make(map[string]CommandHandler) // Moved declaration to the global scope

func GetBot() *Bot {
	return instance
}

// File: internal/bot/bot.go

func NewBot(config *Config, logger *logrus.Logger) (*Bot, error) {
	ctx, cancel := context.WithCancel(context.Background())
	db, err := InitDB(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Log the salt value to verify it is loaded correctly
	logger.Debugf("Loaded Gift Code Salt: %s", config.GiftCode.Salt)

	bot := &Bot{
		Config:          config,
		DB:              db,
		logger:          logger,
		HandlerRegistry: make(map[string]CommandHandler),
		ctx:             ctx,
		cancel:          cancel,
	}
	if config.Discord.Enabled {
		session, err := discordgo.New("Bot " + config.Discord.Token)
		if err != nil {
			return nil, fmt.Errorf("error creating Discord session: %w", err)
		}
		bot.Session = session
	}

	// Set the base URL for GiftCode after bot creation
	SetGiftCodeBaseURL(config)

	// Process pending registrations
	bot.ProcessPendingRegistrations()

	if err := bot.loadHandlers(); err != nil {
		return nil, fmt.Errorf("failed to load handlers: %w", err)
	}
	instance = bot
	return bot, nil
}

func (b *Bot) ProcessPendingRegistrations() {
	b.GetLogger().Infof("Processing %d pending handler registrations", len(pendingHandlers))
	for name, handler := range pendingHandlers {
		b.RegisterHandler(name, handler)
		b.GetLogger().Infof("Registered handler: %s", name)
	}
	b.GetLogger().Infof("HandlerRegistry now contains %d handlers", len(b.HandlerRegistry))
	// Clear the pending handlers
	pendingHandlers = make(map[string]CommandHandler)
}

func RegisterHandlerLater(name string, handler CommandHandler) {
	pendingHandlers[name] = handler
	logrus.Infof("Handler %s queued for registration", name)
}

func (b *Bot) RegisterHandler(name string, handler CommandHandler) {
	b.HandlerRegistry[name] = handler
	b.GetLogger().Debugf("Registered handler: %s", name)
}

func (b *Bot) Start() error {
	if b.Config.Discord.Enabled {
		if err := InitDiscord(b.Config.Discord.Token, b.GetLogger()); err != nil {
			return fmt.Errorf("failed to initialize Discord: %w", err)
		}
	}
	if err := LoadCommands(b.Config.Paths.CommandsConfig, b.GetLogger(), b.HandlerRegistry); err != nil {
		return fmt.Errorf("failed to load commands: %w", err)
	}
	b.GetLogger().Info("Bot has been started")
	return nil
}

func (b *Bot) loadHandlers() error {
	handlersDir := "./internal/bot/handlers"
	files, err := filepath.Glob(filepath.Join(handlersDir, "*.go"))
	if err != nil {
		return fmt.Errorf("failed to read handlers directory: %w", err)
	}
	for _, file := range files {
		b.GetLogger().Infof("Loading handler file: %s", file)
	}
	// The actual loading of handlers now happens through the pending registrations
	// processed in NewBot, so we don't need to do anything else here.
	return nil
}

func (b *Bot) GetHandlerRegistry() map[string]CommandHandler {
	return b.HandlerRegistry
}

func (b *Bot) LoadCommands(configPath string) error {
	return LoadCommands(configPath, b.GetLogger(), b.HandlerRegistry)
}

func (b *Bot) Shutdown() error {
	b.cancel() // Cancel the context to stop all goroutines
	if b.Config.Discord.Enabled {
		if err := b.Session.Close(); err != nil {
			b.GetLogger().WithError(err).Error("Error closing Discord session")
		}
	}
	if sqlDB, err := b.DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			b.GetLogger().Errorf("Error closing database connection: %v", err)
		} else {
			b.GetLogger().Info("Database connection closed successfully")
		}
	}
	b.GetLogger().Info("Bot has been shut down")
	return nil
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	b.GetLogger().Debugf("Received message: %s from user: %s", m.Content, m.Author.Username)
	err := LoadCommands(b.Config.Paths.CommandsConfig, b.GetLogger(), b.HandlerRegistry)
	if err != nil {
		b.GetLogger().Errorf("Failed to load command config: %v", err)
		return
	}
	HandleCommand(s, m, b.Config)
}

func (b *Bot) IsAdmin(s *discordgo.Session, guildID, userID string) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		b.GetLogger().Errorf("Error fetching guild member: %v", err)
		return false
	}
	for _, roleID := range member.Roles {
		if roleID == b.Config.Discord.RoleID {
			return true
		}
	}
	return false
}

func (b *Bot) SendMessage(s *discordgo.Session, channelID, content string) error {
	_, err := s.ChannelMessageSend(channelID, content)
	return err
}

func (b *Bot) GetLogger() *logrus.Logger {
	return b.logger
}

/*
// Session Management Upgrade (Commented Out)
func (b *Bot) SetSession(session *discordgo.Session) {
	b.Session = session
}

func (b *Bot) GetSession() *discordgo.Session {
	return b.Session
}
*/
