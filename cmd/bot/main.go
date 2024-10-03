package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/greysquirr3l/keeper-app/internal/bot"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	// Initialize logger
	setupLogger()

	// Load configuration
	config, err := bot.LoadConfig("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create a new bot instance
	keeperBot, err := bot.NewBot(config)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server (runs independently)
	server := startHTTPServer(ctx, config.Port, keeperBot)

	// Start the bot in a separate goroutine, Discord is optional
	go func() {
		err := keeperBot.Start(ctx)
		if err != nil {
			log.WithError(err).Error("Failed to start bot")
			// You can continue without Discord functionality but keep the HTTP server running.
		}
	}()

	// Wait for shutdown signal
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal

	// Graceful shutdown
	return shutdown(ctx, server, keeperBot)
}

func setupLogger() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
}

func startHTTPServer(ctx context.Context, port string, bot *bot.Bot) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/healthz", bot.HealthCheckHandler())
	mux.HandleFunc("/oauth2/callback", bot.HandleOAuth2Callback) // Handle OAuth2 callback

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

func shutdown(ctx context.Context, server *http.Server, bot *bot.Bot) error {
	log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	if err := bot.Shutdown(); err != nil {
		log.WithError(err).Error("Error shutting down bot")
	}

	log.Info("Server exiting")
	return nil
}
