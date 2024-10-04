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

func GetConfig() *Config {
	return config
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
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
	if discordRedirectURL := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); discordRedirectURL != "" {
		config.Discord.RedirectURL = discordRedirectURL + "/oauth2/callback"
	}

	// Set the database path
	volumeMountPath := os.Getenv("RAILWAY_VOLUME_MOUNT_PATH")
	if volumeMountPath == "" {
		volumeMountPath = "." // Use current directory if not set
	}
	config.Database.Path = filepath.Join(volumeMountPath, "the_keeper.db")

	return config, nil
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
