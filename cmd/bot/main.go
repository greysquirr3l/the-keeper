package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"the-keeper/internal/bot" // Replace with your actual import path

	"github.com/sirupsen/logrus"
)

func main() {
	config, err := bot.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logger := bot.InitializeLogger(config)

	logger.Debug("Logger initialized with level:", logger.GetLevel())

	bot.SetCommandLogger(logger)
	bot.SetUtilLogger(logger)

	// Check if commands.yaml exists
	commandsYamlPath := filepath.Join("configs", "commands.yaml")
	if _, err := os.Stat(commandsYamlPath); os.IsNotExist(err) {
		logger.Fatalf("commands.yaml not found at %s", commandsYamlPath)
	}

	if err := bot.InitDB(config, logger); err != nil {
		logger.Fatalf("Error initializing database: %v", err)
	}

	bot.RegisterCommands()

	if config.Discord.Enabled {
		logger.Debug("Attempting to initialize Discord bot...")

		err := bot.InitDiscord(config.Discord.Token, logger)
		if err != nil {
			logger.Errorf("Error initializing Discord: %v", err)
			logger.Warn("Continuing without Discord functionality")
		} else {
			logger.Info("Discord bot initialized successfully")
		}
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

	if config.Discord.Enabled {
		if err := bot.CloseDiscord(); err != nil {
			logger.Errorf("Error closing Discord connection: %v", err)
		} else {
			logger.Info("Discord connection closed successfully")
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
