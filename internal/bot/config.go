// config.go
package bot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Discord struct {
		Token         string `mapstructure:"token"`
		ClientID      string `mapstructure:"client_id"`
		ClientSecret  string `mapstructure:"client_secret"`
		RoleID        string `mapstructure:"role_id"`
		RedirectURL   string `mapstructure:"redirect_url"`
		Enabled       bool   `mapstructure:"enabled"`
		CommandPrefix string `mapstructure:"command_prefix"`
	} `mapstructure:"discord"`
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Logging struct {
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"logging"`
	Database struct {
		Path string
	}
	Paths struct {
		CommandsConfig string `mapstructure:"commands_config"`
	} `mapstructure:"paths"`
}

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

	// Set database path
	if dbPath := os.Getenv("RAILWAY_VOLUME_MOUNT_PATH"); dbPath != "" {
		config.Database.Path = filepath.Join(dbPath, "the_keeper.db")
	} else {
		config.Database.Path = "the_keeper.db"
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
