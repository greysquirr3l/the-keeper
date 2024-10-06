// File: internal/bot/bot.go

package bot

import (
	"fmt"
	"path/filepath"
	"plugin"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Bot struct {
	Config          *Config
	Session         *discordgo.Session
	DB              *gorm.DB
	shutdownChan    chan struct{}
	logger          *logrus.Logger
	HandlerRegistry map[string]CommandHandler
}

func NewBot(config *Config, logger *logrus.Logger) (*Bot, error) {
	db, err := InitDB(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	bot := &Bot{
		Config:          config,
		DB:              db,
		shutdownChan:    make(chan struct{}),
		logger:          logger,
		HandlerRegistry: make(map[string]CommandHandler),
	}

	if config.Discord.Enabled {
		session, err := discordgo.New("Bot " + config.Discord.Token)
		if err != nil {
			return nil, fmt.Errorf("error creating Discord session: %w", err)
		}
		bot.Session = session
	}

	if err := bot.loadHandlers(); err != nil {
		return nil, fmt.Errorf("failed to load handlers: %w", err)
	}

	return bot, nil
}

func (b *Bot) Start() error {
	if b.Config.Discord.Enabled {
		if err := InitDiscord(b.Config.Discord.Token, b.logger); err != nil {
			return fmt.Errorf("failed to initialize Discord: %w", err)
		}
	}

	if err := LoadCommands(b.Config.Paths.CommandsConfig, b.logger, b.HandlerRegistry); err != nil {
		return fmt.Errorf("failed to load commands: %w", err)
	}

	b.logger.Info("Bot has been started")
	return nil
}

func (b *Bot) loadHandlers() error {
	handlersDir := "./internal/bot/handlers"
	files, err := filepath.Glob(filepath.Join(handlersDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to read handlers directory: %w", err)
	}

	for _, file := range files {
		p, err := plugin.Open(file)
		if err != nil {
			b.logger.Errorf("Failed to open handler %s: %v", file, err)
			continue
		}

		registerSymbol, err := p.Lookup("Register")
		if err != nil {
			b.logger.Errorf("Failed to find Register function in handler %s: %v", file, err)
			continue
		}

		registerFunc, ok := registerSymbol.(func(*Bot))
		if !ok {
			b.logger.Errorf("Invalid Register function signature in handler %s", file)
			continue
		}

		registerFunc(b)
		b.logger.Infof("Successfully loaded handler: %s", file)
	}

	return nil
}

func (b *Bot) RegisterHandler(name string, handler CommandHandler) {
	b.HandlerRegistry[name] = handler
	b.logger.Debugf("Registered handler: %s", name)
}

func (b *Bot) Shutdown() error {
	if b.Config.Discord.Enabled {
		if err := b.Session.Close(); err != nil {
			b.logger.WithError(err).Error("Error closing Discord session")
		}
	}

	if sqlDB, err := b.DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			b.logger.Errorf("Error closing database connection: %v", err)
		} else {
			b.logger.Info("Database connection closed successfully")
		}
	}

	b.logger.Info("Bot has been shut down")
	return nil
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	b.logger.Debugf("Received message: %s from user: %s", m.Content, m.Author.Username)

	err := LoadCommands(b.Config.Paths.CommandsConfig, b.logger, b.HandlerRegistry)
	if err != nil {
		b.logger.Errorf("Failed to load command config: %v", err)
		return
	}

	HandleCommand(s, m, b.Config)
}

func (b *Bot) IsAdmin(s *discordgo.Session, guildID, userID string) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		b.logger.Errorf("Error fetching guild member: %v", err)
		return false
	}

	for _, roleID := range member.Roles {
		if roleID == b.Config.Discord.RoleID {
			return true
		}
	}

	return false
}
