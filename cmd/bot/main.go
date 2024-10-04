package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"the-keeper/internal/bot"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	// Check for the Railway `PORT` environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback to default port if not set
	}

	// Load the bot configuration (YAML)
	config, err := bot.LoadConfig("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load bot configuration: %w", err)
	}

	// Load the commands configuration (YAML)
	commandsConfig, err := bot.LoadCommandsConfig("configs/commands.yaml")
	if err != nil {
		return fmt.Errorf("failed to load commands configuration: %w", err)
	}

	// Create a new bot instance (delegating to bot package)
	keeperBot, err := bot.NewBot(config)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	// Initialize context for the bot and HTTP server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the bot session and register the commands
	if config.Discord.Enabled {
		// Set up a handler for messages
		keeperBot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			// Ignore bot's own messages
			if m.Author.ID == s.State.User.ID {
				return
			}

			// Handle the command with cooldowns
			bot.HandleCommand(s, m, commandsConfig)
		})

		go func() {
			if err := keeperBot.Start(ctx); err != nil {
				log.WithError(err).Error("Failed to start bot")
			}
		}()
	}

	// Start HTTP server and listen for shutdown signal
	server := startHTTPServer(ctx, port, keeperBot)
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal

	return shutdown(ctx, server, keeperBot)
}

func startHTTPServer(ctx context.Context, port string, keeperBot *bot.Bot) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/healthz", keeperBot.HealthCheckHandler())
	mux.HandleFunc("/oauth2/callback", keeperBot.HandleOAuth2Callback)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.WithField("port", port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("HTTP server error")
		}
	}()

	return server
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the Keeper Bot!")
}

func shutdown(ctx context.Context, server *http.Server, keeperBot *bot.Bot) error {
	log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	if err := keeperBot.Shutdown(); err != nil {
		log.WithError(err).Error("Error shutting down bot")
	}

	log.Info("Server exited successfully")
	return nil
}
