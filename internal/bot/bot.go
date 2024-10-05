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
	Config       *Config
	Session      *discordgo.Session
	DB           *gorm.DB
	shutdownChan chan struct{}
	logger       *logrus.Logger
	Managers     map[string]interface{}
}

func NewBot(config *Config, logger *logrus.Logger) (*Bot, error) {
	db, err := InitDB(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	bot := &Bot{
		Config:       config,
		DB:           db,
		shutdownChan: make(chan struct{}),
		logger:       logger,
		Managers:     make(map[string]interface{}),
	}

	if config.Discord.Enabled {
		session, err := discordgo.New("Bot " + config.Discord.Token)
		if err != nil {
			return nil, fmt.Errorf("error creating Discord session: %w", err)
		}
		bot.Session = session
	}

	// Dynamically load managers
	if err := bot.loadManagers(); err != nil {
		return nil, fmt.Errorf("failed to load managers: %w", err)
	}

	// Set the gift code base URL
	SetGiftCodeBaseURL(config)

	return bot, nil
}

func (b *Bot) loadManagers() error {
	managersDir := "./internal/bot/handlers"
	files, err := filepath.Glob(filepath.Join(managersDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to read handlers directory: %w", err)
	}

	for _, file := range files {
		p, err := plugin.Open(file)
		if err != nil {
			b.logger.Errorf("Failed to open handler %s: %v", file, err)
			continue
		}

		newFunc, err := p.Lookup("New")
		if err != nil {
			b.logger.Errorf("Failed to find New function in handler %s: %v", file, err)
			continue
		}

		manager, err := newFunc.(func(*Config, *logrus.Logger) (interface{}, error))(b.Config, b.logger)
		if err != nil {
			b.logger.Errorf("Failed to create manager from handler %s: %v", file, err)
			continue
		}

		managerName := filepath.Base(file[:len(file)-3]) // Remove .so extension
		b.Managers[managerName] = manager
		b.logger.Infof("Loaded Handlers: %s", managerName)
	}

	return nil
}

func (b *Bot) Start() error {
	if b.Config.Discord.Enabled {
		if err := b.initDiscord(); err != nil {
			return fmt.Errorf("failed to initialize Discord: %w", err)
		}
	}

	b.logger.Info("Bot has been started")
	return nil
}

func (b *Bot) Shutdown() error {
	if b.Config.Discord.Enabled {
		if err := b.Session.Close(); err != nil {
			b.logger.WithError(err).Error("Error closing Discord session")
		}
	}
	b.logger.Info("Bot has been shut down")
	return nil
}

func (b *Bot) initDiscord() error {
	b.Session.AddHandler(b.messageCreate)
	b.Session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %w", err)
	}

	b.logger.Info("Discord bot is now running")
	return nil
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	b.logger.Debugf("Received message: %s from user: %s", m.Content, m.Author.Username)

	err := LoadCommands(b.Config.Paths.CommandsConfig)
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
