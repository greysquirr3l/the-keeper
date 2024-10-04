package bot

import (
	"fmt"
	"net/url"
	"os"
)

// getOAuth2URL generates the Discord OAuth2 authorization URL
func getOAuth2URL() string {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	redirectURI := "https://balanced-clarity-production.up.railway.app/oauth2/callback"
	scopes := "identify guilds bot"

	// Return the OAuth2 URL formatted with the client ID, redirect URI, and scopes
	return fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		clientID, redirectURI, url.QueryEscape(scopes),
	)
}

// Initialize the bot session and log with Logrus
// func InitDiscordSession(token string) (*discordgo.Session, error) {
//	session, err := discordgo.New("Bot " + token)
//	if err != nil {
//		logrus.WithError(err).Error("Failed to create Discord session")
//		return nil, err
//	}

//	// Set up intents
//	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

//	return session, nil
//}
// Usage
//package main

//import (
//	"fmt"
//	"the-keeper/internal/bot"
//)

//func main() {
//	// Example of printing the OAuth2 URL
//	fmt.Println("Discord OAuth2 Authorization URL:", bot.getOAuth2URL())
//}

// Other Discord bot logic (e.g., InitDiscordSession, StartDiscordBot)
