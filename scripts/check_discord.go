package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if available
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Continuing with environment variables.")
	}

	// Fetch the Discord Bot Token from the environment variables
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN is not set in the environment.")
	}

	// Create a new Discord session using the bot token
	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Open a connection to Discord
	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %v", err)
	}

	log.Println("Successfully connected to Discord!")

	// Close the Discord session after testing
	defer session.Close()

	// Keep the session running for a short period to test the connection
	log.Println("Bot is running. Press CTRL+C to exit.")
	select {} // Block to keep the program running
}
