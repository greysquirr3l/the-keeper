// cmd/bot/main.go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/greysquirr3l/keeper-bot/internal/bot"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	// Load config
	config, err := bot.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up HTTP server
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/oauth2/callback", oauth2CallbackHandler)

	go func() {
		log.Infof("Starting HTTP server on port %s", config.Server.Port)
		if err := http.ListenAndServe(":"+config.Server.Port, nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start Discord Bot
	if err := startDiscordBot(config); err != nil {
		log.Fatalf("Failed to start Discord bot: %v", err)
	}

	// Graceful shutdown on interrupt signal
	quit := make(chan os.Signal, 1)
	<-quit
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from the Keeper Bot!")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Healthy!")
}

func oauth2CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Received code: %s", code)
	// Here you would exchange the code for a token using the OAuth2 flow
}

func startDiscordBot(cfg *bot.Config) error {
	discord, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return fmt.Errorf("Error creating Discord session: %v", err)
	}

	discord.AddHandler(messageCreate)
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = discord.Open()
	if err != nil {
		return fmt.Errorf("Error opening Discord session: %v", err)
	}

	log.Info("Bot is now running. Press CTRL+C to exit.")
	return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}
