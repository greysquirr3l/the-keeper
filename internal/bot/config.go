package bot

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

type Bot struct {
	Config       *Config
	Session      *discordgo.Session
	shutdownChan chan struct{}
}

type Config struct {
	Discord struct {
		Token         string `yaml:"token"`
		ClientID      string `yaml:"client_id"`
		ClientSecret  string `yaml:"client_secret"`
		RedirectURL   string `yaml:"redirect_url"`
		Enabled       bool   `yaml:"enabled"`        // Add Enabled to config parsing
		CommandPrefix string `yaml:"command_prefix"` // Add Command Prefix for bot commands
	} `yaml:"discord"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Logging struct {
		LogLevel string `yaml:"log_level"` // Add LogLevel for setting log level
	} `yaml:"logging"`
}

type Subcommand struct {
	Description string `yaml:"description"`
	Usage       string `yaml:"usage"`
	Cooldown    string `yaml:"cooldown"`
	Hidden      bool   `yaml:"hidden"`
}

type Command struct {
	Description string                `yaml:"description"`
	Usage       string                `yaml:"usage"`
	Cooldown    string                `yaml:"cooldown"`
	Hidden      bool                  `yaml:"hidden"` // Add Hidden field for main commands
	Subcommands map[string]Subcommand `yaml:"subcommands"`
}

type CommandConfig struct {
	Prefix   string             `yaml:"prefix"`
	Commands map[string]Command `yaml:"commands"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var cfg Config
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
