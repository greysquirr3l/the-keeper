// File: internal/bot/bot.go

package bot

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Bot struct {
	Config           *Config
	Session          *discordgo.Session
	DB               *gorm.DB
	logger           *logrus.Logger
	HandlerRegistry  map[string]CommandHandler
	ctx              context.Context
	cancel           context.CancelFunc
	lastCheckedCodes []GiftCode
	scrapeMutex      sync.Mutex
	Code             string
	Description      string
	Source           string
	Logger           *logrus.Logger
}

var instance *Bot

func GetBot() *Bot {
	return instance
}

func NewBot(config *Config, logger *logrus.Logger) (*Bot, error) {
	ctx, cancel := context.WithCancel(context.Background())
	db, err := InitDB(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
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

	// Process pending registrations
	bot.ProcessPendingRegistrations()

	if err := bot.loadHandlers(); err != nil {
		return nil, fmt.Errorf("failed to load handlers: %w", err)
	}
	instance = bot
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
	files, err := filepath.Glob(filepath.Join(handlersDir, "*.go"))
	if err != nil {
		return fmt.Errorf("failed to read handlers directory: %w", err)
	}
	for _, file := range files {
		b.logger.Infof("Loading handler file: %s", file)
	}
	// The actual loading of handlers now happens through the pending registrations
	// processed in NewBot, so we don't need to do anything else here.
	return nil
}

var pendingHandlers = make(map[string]CommandHandler)

func (b *Bot) RegisterHandler(name string, handler CommandHandler) {
	b.HandlerRegistry[name] = handler
	b.logger.Debugf("Registered handler: %s", name)
}

func RegisterHandlerLater(name string, handler CommandHandler) {
	pendingHandlers[name] = handler
	logrus.Infof("Handler %s queued for registration", name)
}

func (b *Bot) ProcessPendingRegistrations() {
	b.logger.Infof("Processing %d pending handler registrations", len(pendingHandlers))
	for name, handler := range pendingHandlers {
		b.RegisterHandler(name, handler)
		b.logger.Infof("Registered handler: %s", name)
	}
	b.logger.Infof("HandlerRegistry now contains %d handlers", len(b.HandlerRegistry))
	// Clear the pending handlers
	pendingHandlers = make(map[string]CommandHandler)
}

func (b *Bot) GetHandlerRegistry() map[string]CommandHandler {
	return b.HandlerRegistry
}

func (b *Bot) LoadCommands(configPath string) error {
	return LoadCommands(configPath, b.logger, b.HandlerRegistry)
}

func (b *Bot) Shutdown() error {
	b.cancel() // Cancel the context to stop all goroutines
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

func (b *Bot) GetAllGiftCodeRedemptionsPaginated(page, itemsPerPage int) ([]GiftCodeRedemption, error) {
	var redemptions []GiftCodeRedemption
	offset := (page - 1) * itemsPerPage
	result := b.DB.Order("redeemed_at desc").Offset(offset).Limit(itemsPerPage).Find(&redemptions)
	return redemptions, result.Error
}

func (b *Bot) GetUserGiftCodeRedemptionsPaginated(discordID string, page, itemsPerPage int) ([]GiftCodeRedemption, error) {
	var redemptions []GiftCodeRedemption
	offset := (page - 1) * itemsPerPage
	result := b.DB.Where("discord_id = ?", discordID).Order("redeemed_at desc").Offset(offset).Limit(itemsPerPage).Find(&redemptions)
	return redemptions, result.Error
}

func (b *Bot) GetPlayerID(discordID string) (string, error) {
	return GetPlayerID(discordID)
}

func (b *Bot) RecordGiftCodeRedemption(discordID, playerID, giftCode, status string) error {
	return RecordGiftCodeRedemption(discordID, playerID, giftCode, status)
}

func (b *Bot) GetAllPlayerIDs() (map[string]string, error) {
	return GetAllPlayerIDs()
}

func (b *Bot) SendMessage(s *discordgo.Session, channelID, content string) error {
	_, err := s.ChannelMessageSend(channelID, content)
	return err
}

func (b *Bot) GetLogger() *logrus.Logger {
	return b.logger
}

type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate, []string, *Command)
