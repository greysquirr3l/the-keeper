// cmd/bot/main.go
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

	"the-keeper/internal/bot"
)

func main() {
	if err := run(); err != nil {
		bot.Log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	// Define command-line flags
	var port string
	flag.StringVar(&port, "port", "8080", "HTTP server port")
	flag.Parse()

	// Load configuration (YAML)
	config, err := bot.LoadConfig("configs/config.template.yaml") // Change to template if needed
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize the logger using the configuration
	bot.InitializeLogger(config)

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
				bot.Log.WithError(err).Error("Failed to start Discord bot")
			}
		}()
	} else {
		bot.Log.Warn("Discord is disabled in the configuration.")
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
		bot.Log.WithField("port", port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			bot.Log.WithError(err).Error("HTTP server error")
		}
	}()

	return server
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the Keeper Bot!")
}

func shutdown(ctx context.Context, server *http.Server, keeperBot *bot.Bot) error {
	bot.Log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		bot.Log.WithError(err).Error("Server forced to shutdown")
	}

	if err := keeperBot.Shutdown(); err != nil {
		bot.Log.WithError(err).Error("Error shutting down bot")
	}

	bot.Log.Info("Server exited successfully")
	return nil
}
