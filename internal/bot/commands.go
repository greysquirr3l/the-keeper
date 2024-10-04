// commands.go
package bot

import (
	"fmt"
	"io/ioutil"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	commandHandlers    = make(map[string]CommandHandler)
	subcommandHandlers = make(map[string]map[string]CommandHandler)
	cmdLogger          *logrus.Logger
)

type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate, []string)

type CommandConfig struct {
	Prefix   string             `yaml:"prefix"`
	Commands map[string]Command `yaml:"commands"`
}

type Command struct {
	Description string                `yaml:"description"`
	Usage       string                `yaml:"usage"`
	Cooldown    string                `yaml:"cooldown"`
	Subcommands map[string]Subcommand `yaml:"subcommands"`
	Hidden      bool                  `yaml:"hidden"`
}

type Subcommand struct {
	Description string `yaml:"description"`
	Usage       string `yaml:"usage"`
	Cooldown    string `yaml:"cooldown"`
	Hidden      bool   `yaml:"hidden"`
}

func LoadCommandConfig(filename string) (*CommandConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading command config file: %w", err)
	}

	var config CommandConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling command config: %w", err)
	}

	return &config, nil
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, config *CommandConfig) {
	// ... (existing HandleCommand function remains the same)
}

func RegisterCommand(name string, handler CommandHandler) {
	commandHandlers[name] = handler
}

func RegisterSubcommand(cmdName, subCmdName string, handler CommandHandler) {
	if _, ok := subcommandHandlers[cmdName]; !ok {
		subcommandHandlers[cmdName] = make(map[string]CommandHandler)
	}
	subcommandHandlers[cmdName][subCmdName] = handler
}

func SetCommandLogger(logger *logrus.Logger) {
	cmdLogger = logger
}
