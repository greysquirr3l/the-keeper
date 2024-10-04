// main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"the-keeper/internal/bot" // Replace with your actual import path

	"github.com/sirupsen/logrus"
)

func main() {
	config, err := bot.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	logger := bot.InitializeLogger(config)

	// Set loggers for different packages
	bot.SetCommandLogger(logger)
	bot.SetUtilLogger(logger)

	if err := bot.InitDB(config, logger); err != nil {
		logger.Fatalf("Error initializing database: %v", err)
	}

	if config.Discord.Enabled {
		if err := bot.InitDiscord(config.Discord.Token, logger); err != nil {
			logger.Errorf("Error initializing Discord: %v", err)
		}
	}

	// Set up HTTP server
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/healthz", handleHealthCheck)
	http.HandleFunc("/oauth2/callback", handleOAuth2Callback(logger))

	go func() {
		if err := http.ListenAndServe(":"+config.Server.Port, nil); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Infof("Server is running on port %s", config.Server.Port)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	if config.Discord.Enabled {
		if err := bot.CloseDiscord(); err != nil {
			logger.Errorf("Error closing Discord connection: %v", err)
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
		}).Info("Received OAuth2 callback")

		// You can add more processing here if needed

		// Respond to the client
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OAuth2 callback received"))
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
		logger.Infof("Received request: %s", string(jsonDetails))
	}
}
