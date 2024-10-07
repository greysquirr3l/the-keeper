// File: cmd/bot/main.go

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"the-keeper/internal/bot"
	_ "the-keeper/internal/bot/handlers" // Import handlers for deferred registration

	"github.com/sirupsen/logrus"
)

func listDirectoryContents(path string, logger *logrus.Logger) {
	files, err := os.ReadDir(path)
	if err != nil {
		logger.Errorf("Failed to read directory %s: %v", path, err)
		return
	}

	logger.Debugf("Contents of %s:", path)
	for _, file := range files {
		logger.Debugf("- %s", file.Name())
	}
}

func checkConfiguration(config *bot.Config, logger *logrus.Logger) error {
	if config.Discord.Token == "" {
		return fmt.Errorf("Discord token is not set")
	}
	logger.WithFields(logrus.Fields{
		"APIEndpoint": config.GiftCode.APIEndpoint,
		"MinLength":   config.GiftCode.MinLength,
		"MaxLength":   config.GiftCode.MaxLength,
	}).Info("Gift Code Configuration")
	if config.GiftCode.APIEndpoint == "" {
		return fmt.Errorf("Gift code API endpoint is not set")
	}
	return nil
}

func performStartupChecks(b *bot.Bot) error {
	sqlDB, err := b.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	if _, err := b.Session.User("@me"); err != nil {
		return fmt.Errorf("Discord connection failed: %w", err)
	}

	return nil
}

func main() {
	config, err := bot.LoadConfig()
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}

	// Add this line
	fmt.Printf("Main: Discord Role ID from config: %s\n", config.Discord.RoleID)

	logrus.WithField("config", config).Info("Loaded configuration")

	logger := bot.InitializeLogger(config)

	if err := checkConfiguration(config, logger); err != nil {
		logger.Fatalf("Configuration error: %v", err)
	}

	logger.Debug("Logger initialized with level:", logger.GetLevel())

	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed to get current directory: %v", err)
	} else {
		logger.Debugf("Current working directory: %s", currentDir)
	}

	listDirectoryContents(currentDir, logger)
	listDirectoryContents(filepath.Join(currentDir, "configs"), logger)

	commandsYamlPath := filepath.Join("configs", "commands.yaml")
	if _, err := os.Stat(commandsYamlPath); os.IsNotExist(err) {
		logger.Fatalf("commands.yaml not found at %s", commandsYamlPath)
	}

	// Initialize the bot instance
	discordBot, err := bot.NewBot(config, logger)
	if err != nil {
		logger.Fatalf("Error creating bot: %v", err)
	}

	// Process all pending handler registrations after bot creation
	discordBot.ProcessPendingRegistrations()

	// Load commands after handlers have been registered
	if err := bot.LoadCommands(commandsYamlPath, logger, discordBot.GetHandlerRegistry()); err != nil {
		logger.Fatalf("Error loading commands: %v", err)
	}

	if config.Discord.Enabled {
		logger.Debug("Attempting to initialize Discord bot...")

		if err := performStartupChecks(discordBot); err != nil {
			logger.Fatalf("Startup checks failed: %v", err)
		}

		err = discordBot.Start()
		if err != nil {
			logger.Fatalf("Error starting bot: %v", err)
		}

		logger.Info("Discord bot initialized and started successfully")
	} else {
		logger.Info("Discord bot is disabled in configuration")
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/healthz", handleHealthCheck)
	http.HandleFunc("/oauth2/callback", handleOAuth2Callback(logger))

	go func() {
		logger.Infof("Starting HTTP server on port %s", config.Server.Port)
		if err := http.ListenAndServe(":"+config.Server.Port, nil); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Info("Server is now running. Press CTRL+C to exit.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	if config.Discord.Enabled && discordBot != nil {
		if err := discordBot.Shutdown(); err != nil {
			logger.Errorf("Error shutting down bot: %v", err)
		} else {
			logger.Info("Bot shut down successfully")
		}
	}

	logger.Info("Server stopped")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the bot server!"))
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleOAuth2Callback(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received OAuth2 callback request")

		err := r.ParseForm()
		if err != nil {
			logger.Errorf("Error parsing form: %v", err)
			http.Error(w, "Error processing request", http.StatusBadRequest)
			return
		}

		logger.WithFields(logrus.Fields{
			"code":  r.Form.Get("code"),
			"state": r.Form.Get("state"),
		}).Info("OAuth2 callback parameters")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OAuth2 callback received"))
		logger.Info("OAuth2 callback processed successfully")
	}
}
