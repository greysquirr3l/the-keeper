// File: internal/bot/config.go

package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var config *Config

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("$HOME/.the-keeper")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found. Using default values and environment variables.")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	config = &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Parse API timeout
	if config.GiftCode.APITimeout == 0 {
		config.GiftCode.APITimeout = 30 * time.Second // Default to 30 seconds if not specified
	} else {
		duration, err := time.ParseDuration(fmt.Sprintf("%ds", config.GiftCode.APITimeout))
		if err != nil {
			return nil, fmt.Errorf("invalid API timeout: %w", err)
		}
		config.GiftCode.APITimeout = duration
	}

	// Override with environment variables if present
	if discordToken := os.Getenv("DISCORD_BOT_TOKEN"); discordToken != "" {
		config.Discord.Token = discordToken
		fmt.Println("Discord Bot Token set from environment")
	}
	if discordClientID := os.Getenv("DISCORD_CLIENT_ID"); discordClientID != "" {
		config.Discord.ClientID = discordClientID
		fmt.Println("Discord Client ID set from environment")
	}
	if discordClientSecret := os.Getenv("DISCORD_CLIENT_SECRET"); discordClientSecret != "" {
		config.Discord.ClientSecret = discordClientSecret
		fmt.Println("Discord Client Secret set from environment")
	}
	if discordRoleID := os.Getenv("DISCORD_ROLE_ID"); discordRoleID != "" {
		config.Discord.RoleID = discordRoleID
		fmt.Printf("Discord Role ID set from environment: %s\n", discordRoleID)
	} else {
		fmt.Println("DISCORD_ROLE_ID not set in environment")
	}
	if redirectURL := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); redirectURL != "" {
		config.Discord.RedirectURL = redirectURL + "/oauth2/callback"
		fmt.Printf("Redirect URL set from environment: %s\n", config.Discord.RedirectURL)
	}
	// Set database path
	if dbPath := os.Getenv("RAILWAY_VOLUME_MOUNT_PATH"); dbPath != "" {
		config.Database.Path = filepath.Join(dbPath, config.Database.Name)
	} else {
		config.Database.Path = config.Database.Name
	}
	if notificationChannelID := os.Getenv("DISCORD_NOTIFICATION_CHANNEL_ID"); notificationChannelID != "" {
		config.Discord.NotificationChannelID = notificationChannelID
	}
	if giftCodeSalt := os.Getenv("GIFT_CODE_SALT"); giftCodeSalt != "" {
		config.GiftCode.Salt = giftCodeSalt
		fmt.Printf("Gift Code Salt set from environment: %s\n", config.GiftCode.Salt)
	} else {
		fmt.Println("GIFT_CODE_SALT not set in environment")
	}

	// Set default values if not provided
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Logging.LogLevel == "" {
		config.Logging.LogLevel = "info"
	}
	if config.Discord.CommandPrefix == "" {
		config.Discord.CommandPrefix = "!"
	}
	if config.Paths.CommandsConfig == "" {
		config.Paths.CommandsConfig = "configs/commands.yaml"
	}

	logrus.WithFields(logrus.Fields{
		"DiscordEnabled": config.Discord.Enabled,
		"ServerPort":     config.Server.Port,
		"LogLevel":       config.Logging.LogLevel,
		"GiftCodeAPI":    config.GiftCode.APIEndpoint,
		"GiftCodeSalt":   config.GiftCode.Salt,
	}).Info("Loaded configuration")

	fmt.Printf("Loaded Gift Code Salt: %s\n", config.GiftCode.Salt)
	fmt.Printf("Final Discord Role ID: %s\n", config.Discord.RoleID)

	return config, nil
}

func GetConfig() *Config {
	return config
}

func InitializeLogger(config *Config) *logrus.Logger {
	log := logrus.New()
	level, err := logrus.ParseLevel(config.Logging.LogLevel)
	if err != nil {
		level = logrus.InfoLevel // Default to info if parsing fails
	}
	log.SetLevel(level)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return log
}
