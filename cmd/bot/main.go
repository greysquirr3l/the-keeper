package main

import (
	"context"
	"flag"
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
	// Define command-line flags
	var port string
	flag.StringVar(&port, "port", "8080", "HTTP server port")
	flag.Parse()

	// Load configuration (YAML)
	config, err := bot.LoadConfig("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create a new bot instance (delegating to bot package)
	keeperBot, err := bot.NewBot(config)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	// Start the HTTP server in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := startHTTPServer(ctx, port, keeperBot)

	// Optionally start the bot in a separate goroutine if Discord is enabled
	if config.Discord.Enabled {
		go func() {
			if err := keeperBot.Start(ctx); err != nil {
				log.WithError(err).Error("Failed to start bot")
			}
		}()
	}

	// Listen for system signals (e.g., SIGINT, SIGTERM)
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal

	// Graceful shutdown
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
