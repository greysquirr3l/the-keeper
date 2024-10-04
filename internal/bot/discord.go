package bot

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// StartDiscordBot sets up the Discord message handler and starts the bot session.
func StartDiscordBot(ctx context.Context, keeperBot *Bot, commandsConfig *CommandConfig) {
	if keeperBot.Config.Discord.Enabled {
		// Set up a handler for messages
		keeperBot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			// Ignore bot's own messages
			if m.Author.ID == s.State.User.ID {
				return
			}

			// Log the received message for debugging
			log.Debugf("Message received from user: %s, content: %s", m.Author.Username, m.Content)

			// Handle the command with cooldowns
			HandleCommand(s, m, commandsConfig)
		})

		go func() {
			if err := keeperBot.Start(ctx); err != nil {
				log.WithError(err).Error("Failed to start Discord bot")
			}
		}()

		log.Info("Discord bot has been successfully initialized and started.")
	} else {
		log.Warn("Discord is disabled in the configuration.")
	}
}

// InitDiscordSession initializes a new Discord session with the required intents and logging.
func InitDiscordSession(token string) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Error("Failed to create Discord session")
		return nil, err
	}

	// Set the necessary intents for the bot
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// Enable logging for certain events
	session.AddHandler(func(s *discordgo.Session, event *discordgo.Ready) {
		log.Info("Bot is ready and connected to Discord.")
	})

	session.AddHandler(func(s *discordgo.Session, event *discordgo.Disconnect) {
		log.Warn("Bot disconnected from Discord.")
	})

	log.Info("Discord session initialized successfully")
	return session, nil
}
