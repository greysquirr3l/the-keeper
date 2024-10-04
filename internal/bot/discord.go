// discord.go
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
	var err error
	discordSession, err = discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}

	// Set the intents
	discordSession.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	discordSession.AddHandler(messageCreate)

	err = discordSession.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	discordLogger.Info("Discord bot is now running")
	return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	config := GetConfig()
	commandConfig, err := LoadCommandConfig(config.Paths.CommandsConfig)
	if err != nil {
		discordLogger.Errorf("Failed to load command config: %v", err)
		return
	}

	HandleCommand(s, m, commandConfig)
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
		return discordSession.Close()
	}
	return nil
}
