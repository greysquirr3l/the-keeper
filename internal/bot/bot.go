// bot.go
package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	Config       *Config
	Session      *discordgo.Session
	shutdownChan chan struct{}
	logger       *logrus.Logger
}

func NewBot(config *Config, logger *logrus.Logger) (*Bot, error) {
	bot := &Bot{
		Config:       config,
		shutdownChan: make(chan struct{}),
		logger:       logger,
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

func (b *Bot) Start() error {
	if b.Config.Discord.Enabled {
		b.logger.Info("Starting Discord bot")
		err := b.initDiscord()
		if err != nil {
			return fmt.Errorf("failed to initialize Discord: %w", err)
		}
	}
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
	// TODO: err = LoadCommands(b.Config.Paths.CommandsConfig) err := LoadCommands(b.Config.Paths.CommandsConfig)

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
