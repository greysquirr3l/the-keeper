package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var (
	discordSession *discordgo.Session
	discordLogger  *logrus.Logger
)

func InitDiscord(token string, logger *logrus.Logger) error {
	discordLogger = logger
	discordLogger.Info("Initializing Discord bot...")

	// Enable discordgo debug logging
	discordgo.Logger = func(msgL, caller int, format string, a ...interface{}) {
		discordLogger.Debugf(format, a...)
	}

	var err error
	discordLogger.Info("Creating new Discord session...")
	discordSession, err = discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}

	// Set the log level for discordgo to debug
	discordSession.LogLevel = discordgo.LogDebug

	discordLogger.Info("Setting up intents...")
	discordSession.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	discordLogger.Info("Adding message handler...")
	discordSession.AddHandler(messageCreate)

	// Add connect and disconnect handlers
	discordSession.AddHandler(func(s *discordgo.Session, _ *discordgo.Connect) {
		discordLogger.Info("Bot has connected to Discord")
	})
	discordSession.AddHandler(func(s *discordgo.Session, _ *discordgo.Disconnect) {
		discordLogger.Warn("Bot has disconnected from Discord")
	})

	discordLogger.Info("Opening Discord connection...")
	err = discordSession.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	// Set a custom status
	err = discordSession.UpdateGameStatus(0, "Ready to serve!")
	if err != nil {
		discordLogger.Errorf("Error setting presence: %v", err)
	}

	discordLogger.Infof("Discord bot is now running with username: %s and ID: %s",
		discordSession.State.User.Username,
		discordSession.State.User.ID)
	return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	discordLogger.Debugf("Received message: %s from user: %s", m.Content, m.Author.Username)

	config := GetConfig()
	err := LoadCommands(config.Paths.CommandsConfig)
	if err != nil {
		discordLogger.Errorf("Failed to load command config: %v", err)
		return
	}

	HandleCommand(s, m, config)
}

// SendMessage is a helper function to send a message to a channel
func SendMessage(s *discordgo.Session, channelID string, message string) {
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		logrus.Errorf("Error sending message: %v", err)
	}
}

// Helper function to check if a user has a specific role
func HasRole(s *discordgo.Session, guildID string, userID string, roleID string) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		discordLogger.Errorf("Error fetching guild member: %v", err)
		return false
	}

	for _, r := range member.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}

func IsAdmin(s *discordgo.Session, guildID, userID string) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		discordLogger.Errorf("Error fetching guild member: %v", err)
		return false
	}

	for _, roleID := range member.Roles {
		if roleID == GetConfig().Discord.RoleID {
			return true
		}
	}

	return false
}

func CloseDiscord() error {
	if discordSession != nil {
		discordLogger.Info("Closing Discord session...")
		return discordSession.Close()
	}
	return nil
}
