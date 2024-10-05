// config.go
package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// File: internal/bot/config.go

type Config struct {
	Discord struct {
		Token         string `yaml:"token"`
		ClientID      string `yaml:"client_id"`
		ClientSecret  string `yaml:"client_secret"`
		RoleID        string `yaml:"RoleID"`
		RedirectURL   string `yaml:"redirect_url"`
		Enabled       bool   `yaml:"enabled"`
		CommandPrefix string `yaml:"command_prefix"`
	} `yaml:"discord"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Logging struct {
		LogLevel string `yaml:"log_level"`
	} `yaml:"logging"`
	Paths struct {
		CommandsConfig string `yaml:"commands_config"`
	} `yaml:"paths"`
	Database struct {
		VolumeMountPath string `yaml:"volumeMountPath"`
		Name            string `yaml:"name"`
		Path            string `yaml:"path"`
	} `yaml:"database"`
	GiftCode struct {
		Salt        string        `yaml:"salt"`
		MinLength   int           `yaml:"min_length"`
		MaxLength   int           `yaml:"max_length"`
		APIEndpoint string        `yaml:"api_endpoint"`
		APITimeout  time.Duration `yaml:"api_timeout"`
	} `yaml:"gift_code"`
}

var config *Config

// TODO: Update other files to use viper.

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
	}
	if discordClientID := os.Getenv("DISCORD_CLIENT_ID"); discordClientID != "" {
		config.Discord.ClientID = discordClientID
	}
	if discordClientSecret := os.Getenv("DISCORD_CLIENT_SECRET"); discordClientSecret != "" {
		config.Discord.ClientSecret = discordClientSecret
	}
	if discordRoleID := os.Getenv("DISCORD_ROLE_ID"); discordRoleID != "" {
		config.Discord.RoleID = discordRoleID
	}
	if redirectURL := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); redirectURL != "" {
		config.Discord.RedirectURL = redirectURL + "/oauth2/callback"
	}
	// TODO: update to add giftcode defaults.

	// Set database path
	if dbPath := os.Getenv("RAILWAY_VOLUME_MOUNT_PATH"); dbPath != "" {
		config.Database.Path = filepath.Join(dbPath, config.Database.Name)
	} else {
		config.Database.Path = config.Database.Name
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

	return config, nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	return config
}

func InitializeLogger(config *Config) *logrus.Logger {
	log := logrus.New()
	level, err := logrus.ParseLevel(config.Logging.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return log
}
