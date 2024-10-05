// File: internal/bot/commands.go

package bot

import (
	"fmt"
	"io/ioutil"

	// "os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	// "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate, []string, *Command)

type Command struct {
	Name        string
	Description string
	Usage       string
	Cooldown    string
	Hidden      bool
	Handler     string // Name of the handler function
	Subcommands map[string]*Command
	HandlerFunc CommandHandler // The actual function to be called
}

type CommandConfig struct {
	Prefix   string
	Commands map[string]*Command
}

var (
	CommandRegistry = make(map[string]*Command)
	HandlerRegistry = make(map[string]CommandHandler)
)

var globalLogger *logrus.Logger

func SetLogger(logger *logrus.Logger) {
	globalLogger = logger
}

// func LoadCommands(configPath string, logger *logrus.Logger) error {
func LoadCommands(configPath string) error {
	// // TODO: LoadCommands(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		globalLogger.Errorf("Error reading command config file: %v", err)
		return fmt.Errorf("error reading command config file: %w", err)
	}

	var config CommandConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error unmarshaling command config: %w", err)
	}

	handlersDir := "./bot/handlers" // Set the correct handlers directory path

	for name, cmd := range config.Commands {
		CommandRegistry[name] = cmd

		handlerFileName := fmt.Sprintf("%s_handlers.go", name)
		handlerFilePath := filepath.Join(handlersDir, handlerFileName)

		if fileExists(handlerFilePath) {
			globalLogger.Infof("Loaded handlers for command: %s", name)

			// Check for main command handler
			mainHandlerName := fmt.Sprintf("handle%sCommand", strings.Title(name))
			if handler, exists := HandlerRegistry[mainHandlerName]; exists {
				globalLogger.Infof("  - Registered handler: %s", mainHandlerName)
				cmd.HandlerFunc = handler
			} else {
				globalLogger.Warnf("  - Main handler not found: %s", mainHandlerName)
			}

			// Check for subcommand handlers
			for subCmdName, subCmd := range cmd.Subcommands {
				subHandlerName := fmt.Sprintf("handle%s%sCommand", strings.Title(name), strings.Title(subCmdName))
				if handler, exists := HandlerRegistry[subHandlerName]; exists {
					globalLogger.Infof("  - Registered handler: %s", subHandlerName)
					subCmd.HandlerFunc = handler
				} else {
					globalLogger.Warnf("  - Subcommand handler not found: %s", subHandlerName)
				}
			}
		} else {
			globalLogger.Warnf("Handler file not found for command: %s", name)
		}
	}

	return nil
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, config *Config) {
	args := strings.Fields(m.Content)
	if len(args) == 0 || !strings.HasPrefix(args[0], config.Discord.CommandPrefix) {
		return
	}

	cmdName := strings.TrimPrefix(args[0], config.Discord.CommandPrefix)
	cmd, exists := CommandRegistry[cmdName]
	if !exists {
		SendMessage(s, m.ChannelID, "Unknown command. Use !help to see available commands.")
		return
	}

	if len(args) > 1 && cmd.Subcommands != nil {
		subCmd, subExists := cmd.Subcommands[args[1]]
		if subExists && subCmd.HandlerFunc != nil {
			subCmd.HandlerFunc(s, m, args[2:], subCmd)
			return
		}
	}

	if cmd.HandlerFunc != nil {
		cmd.HandlerFunc(s, m, args[1:], cmd)
	} else {
		SendMessage(s, m.ChannelID, "This command is not implemented.")
	}
}

func RegisterHandler(name string, handler CommandHandler) {
	HandlerRegistry[name] = handler
}
