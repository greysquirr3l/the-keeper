package bot

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
)

type Bot struct {
	Config       *Config
	Session      *discordgo.Session
	shutdownChan chan struct{}
}

// Config structure to load YAML config
type Config struct {
	Port    string
	Discord struct {
		Token         string `yaml:"token"`
		Enabled       bool   `yaml:"enabled"`
		CommandPrefix string `yaml:"command_prefix"`
	} `yaml:"discord"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Logging struct {
		LogLevel string `yaml:"log_level"`
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
func LoadConfig(filename string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	return config, nil
}
