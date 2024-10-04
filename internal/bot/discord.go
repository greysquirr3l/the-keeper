package bot

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// StartDiscordBot sets up the Discord message handler and starts the bot session.
func StartDiscordBot(ctx context.Context, keeperBot *Bot, commandsConfig *CommandConfig) {
	if keeperBot.Config.Discord.Enabled {
		// Set up a handler for messages
		keeperBot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			// Ignore bot's own messages
			if m.Author.ID == s.State.User.ID {
				return
			}

			// Handle the command with cooldowns
			HandleCommand(s, m, commandsConfig)
		})

		go func() {
			if err := keeperBot.Start(ctx); err != nil {
				Log.WithError(err).Error("Failed to start bot")
			}
		}()
	}
}

// InitDiscordSession initializes a new Discord session with the required intents.
func InitDiscordSession(token string) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		Log.WithError(err).Error("Failed to create Discord session")
		return nil, err
	}

	// Set the necessary intents for the bot
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	Log.Info("Discord session initialized successfully")
	return session, nil
}
