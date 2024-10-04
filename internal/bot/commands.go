package bot

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var log = logrus.New()

// Cooldown map to track cooldowns per user and command
var cooldowns = struct {
	sync.Mutex
	commands map[string]map[string]time.Time
}{
	commands: make(map[string]map[string]time.Time),
}

// CommandConfig represents the structure of commands.yaml
type Subcommand struct {
	Description string `yaml:"description"`
	Usage       string `yaml:"usage"`
	Cooldown    string `yaml:"cooldown"` // Cooldown as a string (like "5s")
	Hidden      bool   `yaml:"hidden"`
}

type Command struct {
	Description string                `yaml:"description"`
	Usage       string                `yaml:"usage"`
	Cooldown    string                `yaml:"cooldown"` // Cooldown as a string
	Hidden      bool                  `yaml:"hidden"`
	Subcommands map[string]Subcommand `yaml:"subcommands"`
}

type CommandConfig struct {
	Commands map[string]Command `yaml:"commands"`
}

// LoadCommandsConfig loads the command configuration from a YAML file
func LoadCommandsConfig(filename string) (*CommandConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.WithError(err).Errorf("Failed to open commands config file: %s", filename)
		return nil, fmt.Errorf("failed to open commands config file: %w", err)
	}
	defer file.Close()

	var config CommandConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.WithError(err).Error("Failed to decode commands YAML")
		return nil, fmt.Errorf("failed to decode commands YAML: %w", err)
	}

	return &config, nil
}

// CheckCooldown checks if the cooldown has passed for a command or subcommand
func CheckCooldown(userID, command string, cooldown int) bool {
	if cooldown == 0 {
		return true // No cooldown set for this command
	}

	cooldowns.Lock()
	defer cooldowns.Unlock()

	if cooldowns.commands[command] == nil {
		cooldowns.commands[command] = make(map[string]time.Time)
	}

	lastUsed, exists := cooldowns.commands[command][userID]
	if !exists {
		return true // No cooldown exists for this user
	}

	// Check if enough time has passed since the last use
	if time.Since(lastUsed) >= time.Duration(cooldown)*time.Second {
		return true
	}

	return false
}

// SetCooldown sets the current time as the last used time for the command or subcommand
func SetCooldown(userID, command string) {
	cooldowns.Lock()
	defer cooldowns.Unlock()

	if cooldowns.commands[command] == nil {
		cooldowns.commands[command] = make(map[string]time.Time)
	}

	cooldowns.commands[command][userID] = time.Now()
}

// HandleCommand processes incoming commands and checks for cooldowns dynamically from YAML
func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, commandConfig *CommandConfig) {
	cmdName, subCmdName := parseCommand(m.Content)

	// Find the command in the configuration
	cmd, exists := commandConfig.Commands[cmdName]
	if !exists {
		log.Warnf("Command '%s' not found", cmdName)
		return
	}

	var cooldown time.Duration
	var err error

	if subCmdName != "" {
		subCmd, subExists := cmd.Subcommands[subCmdName]
		if !subExists {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Subcommand '%s' not found.", subCmdName))
			log.Warnf("Subcommand '%s' not found under command '%s'", subCmdName, cmdName)
			return
		}

		// Parse cooldown for subcommand
		cooldown, err = time.ParseDuration(subCmd.Cooldown)
		if err != nil {
			log.Errorf("Invalid cooldown format for subcommand '%s': %v", subCmdName, err)
			return
		}

		// Handle subcommand execution and cooldown
		if !CheckCooldown(m.Author.ID, cmdName+"_"+subCmdName, int(cooldown.Seconds())) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You must wait before using the '%s' subcommand again.", subCmdName))
			return
		}
		executeSubCommand(s, m, subCmd)
		SetCooldown(m.Author.ID, cmdName+"_"+subCmdName)
		log.Infof("Subcommand '%s' executed by user %s", subCmdName, m.Author.Username)
	} else {
		// Parse cooldown for the main command
		cooldown, err = time.ParseDuration(cmd.Cooldown)
		if err != nil {
			log.Errorf("Invalid cooldown format for command '%s': %v", cmdName, err)
			return
		}

		// Handle command execution and cooldown
		if !CheckCooldown(m.Author.ID, cmdName, int(cooldown.Seconds())) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You must wait before using the '%s' command again.", cmdName))
			return
		}
		executeCommand(s, m, cmd)
		SetCooldown(m.Author.ID, cmdName)
		log.Infof("Command '%s' executed by user %s", cmdName, m.Author.Username)
	}
}

// Parse command and subcommand from the message
func parseCommand(message string) (string, string) {
	// Example: "!id add <args>" -> cmdName="id", subCmdName="add"
	parts := strings.Fields(message)
	if len(parts) == 0 || parts[0][0] != '!' {
		return "", ""
	}
	cmdName := parts[0][1:]
	if len(parts) > 1 {
		return cmdName, parts[1] // Return command and subcommand
	}
	return cmdName, "" // No subcommand
}

// Execute the main command (example)
func executeCommand(s *discordgo.Session, m *discordgo.MessageCreate, command Command) {
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Executing command: %s", command.Description))
}

// Execute the subcommand (example)
func executeSubCommand(s *discordgo.Session, m *discordgo.MessageCreate, subCmd Subcommand) {
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Executing subcommand: %s", subCmd.Description))
}
