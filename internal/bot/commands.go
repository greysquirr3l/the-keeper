package bot

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	commandHandlers = make(map[string]CommandHandler)
	cmdLogger       *logrus.Logger
	cooldowns       = make(map[string]time.Time)
	cooldownMutex   sync.Mutex
)

type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate, []string, *Command)

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

func RegisterCommands() {
	RegisterCommand("help", handleHelpCommand)
	RegisterCommand("id", handleGenericCommand)
	RegisterCommand("term", handleGenericCommand)
	RegisterCommand("giftcode", handleGenericCommand)
	RegisterCommand("scrape", handleGenericCommand)
}

func RegisterCommand(name string, handler CommandHandler) {
	commandHandlers[name] = handler
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, config *CommandConfig) {
	if m.Author.Bot {
		return
	}

	content := m.Content
	if !strings.HasPrefix(content, config.Prefix) {
		return
	}

	args := strings.Fields(content[len(config.Prefix):])
	if len(args) == 0 {
		return
	}

	cmdName := args[0]
	cmd, exists := config.Commands[cmdName]
	if !exists {
		cmdLogger.Infof("Unknown command: %s", cmdName)
		return
	}

	if !checkCooldown(m.Author.ID, cmdName, cmd.Cooldown) {
		SendMessage(s, m.ChannelID, "This command is on cooldown. Please wait before using it again.")
		return
	}

	if handler, ok := commandHandlers[cmdName]; ok {
		cmdLogger.WithFields(logrus.Fields{
			"user":    m.Author.Username,
			"command": cmdName,
			"args":    args[1:],
		}).Debug("Executing command")
		handler(s, m, args[1:], &cmd)
	} else {
		cmdLogger.Warnf("Handler not implemented for command: %s", cmdName)
	}
}

func handleGenericCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	if len(args) == 0 {
		SendMessage(s, m.ChannelID, fmt.Sprintf("Usage: %s", cmd.Usage))
		return
	}

	subCmdName := args[0]
	subCmd, exists := cmd.Subcommands[subCmdName]
	if !exists {
		SendMessage(s, m.ChannelID, fmt.Sprintf("Unknown subcommand: %s. Use !help %s for more information.", subCmdName, cmd.Usage))
		return
	}

	if !checkCooldown(m.Author.ID, fmt.Sprintf("%s:%s", cmd.Usage, subCmdName), subCmd.Cooldown) {
		SendMessage(s, m.ChannelID, "This subcommand is on cooldown. Please wait before using it again.")
		return
	}

	// Here you would implement the logic for each subcommand
	// For now, we'll just send a message with the subcommand description
	SendMessage(s, m.ChannelID, fmt.Sprintf("Executing subcommand: %s\nDescription: %s", subCmdName, subCmd.Description))
}

func handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string, cmd *Command) {
	config := GetConfig()
	commandConfig, err := LoadCommandConfig(config.Paths.CommandsConfig)
	if err != nil {
		cmdLogger.Errorf("Failed to load command config: %v", err)
		SendMessage(s, m.ChannelID, "Error loading command configuration.")
		return
	}

	if len(args) == 0 {
		// General help message
		helpMsg := "Available commands:\n"
		for cmdName, cmd := range commandConfig.Commands {
			if !cmd.Hidden {
				helpMsg += fmt.Sprintf("**%s%s**: %s\n", config.Discord.CommandPrefix, cmdName, cmd.Description)
			}
		}
		helpMsg += fmt.Sprintf("\nUse `%shelp <command>` for more information on a specific command.", config.Discord.CommandPrefix)
		SendMessage(s, m.ChannelID, helpMsg)
	} else {
		// Help for specific command
		cmdName := strings.ToLower(args[0])
		cmd, exists := commandConfig.Commands[cmdName]
		if !exists || cmd.Hidden {
			SendMessage(s, m.ChannelID, fmt.Sprintf("No help available for command: %s", cmdName))
			return
		}
		helpMsg := fmt.Sprintf("**%s%s**: %s\n", config.Discord.CommandPrefix, cmdName, cmd.Description)
		helpMsg += fmt.Sprintf("Usage: `%s`\n", cmd.Usage)
		if cmd.Cooldown != "" {
			helpMsg += fmt.Sprintf("Cooldown: %s\n", cmd.Cooldown)
		}
		if len(cmd.Subcommands) > 0 {
			helpMsg += "\nSubcommands:\n"
			for subCmdName, subCmd := range cmd.Subcommands {
				if !subCmd.Hidden {
					helpMsg += fmt.Sprintf("  **%s**: %s\n", subCmdName, subCmd.Description)
					helpMsg += fmt.Sprintf("    Usage: `%s`\n", subCmd.Usage)
					if subCmd.Cooldown != "" {
						helpMsg += fmt.Sprintf("    Cooldown: %s\n", subCmd.Cooldown)
					}
				}
			}
		}
		SendMessage(s, m.ChannelID, helpMsg)
	}
}

func SetCommandLogger(logger *logrus.Logger) {
	cmdLogger = logger
}

func LoadCommandConfig(filename string) (*CommandConfig, error) {
	cmdLogger.Debugf("Attempting to load command config from: %s", filename)

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		cmdLogger.Errorf("Command config file does not exist: %s", filename)
		return nil, fmt.Errorf("command config file does not exist: %s", filename)
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		cmdLogger.Errorf("Error reading command config file: %v", err)
		return nil, fmt.Errorf("error reading command config file: %w", err)
	}

	var config CommandConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		cmdLogger.Errorf("Error unmarshaling command config: %v", err)
		return nil, fmt.Errorf("error unmarshaling command config: %w", err)
	}

	cmdLogger.Debug("Command config loaded successfully")
	return &config, nil
}

func checkCooldown(userID, command, cooldownStr string) bool {
	if cooldownStr == "" {
		return true
	}

	cooldownDuration, err := time.ParseDuration(cooldownStr)
	if err != nil {
		cmdLogger.Errorf("Invalid cooldown duration for command %s: %v", command, err)
		return true
	}

	cooldownMutex.Lock()
	defer cooldownMutex.Unlock()

	key := fmt.Sprintf("%s:%s", userID, command)
	lastUsed, exists := cooldowns[key]
	if !exists || time.Since(lastUsed) > cooldownDuration {
		cooldowns[key] = time.Now()
		return true
	}

	return false
}
