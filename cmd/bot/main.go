// File: cmd/bot/main.go

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"the-keeper/internal/bot"

	"github.com/sirupsen/logrus"
)

func listDirectoryContents(path string, logger *logrus.Logger) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Errorf("Failed to read directory %s: %v", path, err)
		return
	}

	logger.Debugf("Contents of %s:", path)
	for _, file := range files {
		logger.Debugf("- %s", file.Name())
	}
}

func checkConfiguration(config *bot.Config) error {
	if config.Discord.Token == "" {
		return fmt.Errorf("Discord token is not set")
	}
	if config.GiftCode.APIEndpoint == "" {
		return fmt.Errorf("Gift code API endpoint is not set")
	}
	// Add more checks as needed
	return nil
}

func performStartupChecks(b *bot.Bot) error {
	// Check database connection
	sqlDB, err := b.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Check Discord connection
	if _, err := b.Session.User("@me"); err != nil {
		return fmt.Errorf("Discord connection failed: %w", err)
	}

	// Load commands
	if err := bot.LoadCommands(b.Config.Paths.CommandsConfig); err != nil {
		return fmt.Errorf("failed to load commands: %w", err)
	}

	return nil
}

func main() {
	config, err := bot.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := checkConfiguration(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Set the gift code base URL after loading the config
	bot.SetGiftCodeBaseURL(config)

	logger := bot.InitializeLogger(config)

	logger.Debug("Logger initialized with level:", logger.GetLevel())

	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed to get current directory: %v", err)
	} else {
		logger.Debugf("Current working directory: %s", currentDir)
	}

	listDirectoryContents(currentDir, logger)
	listDirectoryContents(filepath.Join(currentDir, "configs"), logger)

	// Check if commands.yaml exists
	commandsYamlPath := filepath.Join("configs", "commands.yaml")
	if _, err := os.Stat(commandsYamlPath); os.IsNotExist(err) {
		logger.Fatalf("commands.yaml not found at %s", commandsYamlPath)
	}

	// Load commands
	err = bot.LoadCommands(commandsYamlPath)
	if err != nil {
		logger.Fatalf("Error loading commands: %v", err)
	}

	var discordBot *bot.Bot
	if config.Discord.Enabled {
		logger.Debug("Attempting to initialize Discord bot...")

		discordBot, err = bot.NewBot(config, logger)
		if err != nil {
			logger.Fatalf("Error creating bot: %v", err)
		}

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

	// Set up HTTP server
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

	// Wait for interrupt signal to gracefully shut down the server
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

		// Log the entire request for debugging purposes
		logRequest(r, logger)

		// Parse the query parameters
		err := r.ParseForm()
		if err != nil {
			logger.Errorf("Error parsing form: %v", err)
			http.Error(w, "Error processing request", http.StatusBadRequest)
			return
		}

		// Log the query parameters
		logger.WithFields(logrus.Fields{
			"code":  r.Form.Get("code"),
			"state": r.Form.Get("state"),
		}).Info("OAuth2 callback parameters")

		// You can add more processing here if needed

		// Respond to the client
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OAuth2 callback received"))
		logger.Info("OAuth2 callback processed successfully")
	}
}

func logRequest(r *http.Request, logger *logrus.Logger) {
	// Create a map to store request details
	requestDetails := map[string]interface{}{
		"Method":     r.Method,
		"RequestURI": r.RequestURI,
		"RemoteAddr": r.RemoteAddr,
		"Header":     r.Header,
	}

	// Log the request details as JSON
	jsonDetails, err := json.MarshalIndent(requestDetails, "", "  ")
	if err != nil {
		logger.Errorf("Error marshaling request details: %v", err)
	} else {
		logger.Infof("Received request details: %s", string(jsonDetails))
	}
}
