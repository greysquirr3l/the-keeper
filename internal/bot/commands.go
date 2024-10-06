// File: internal/bot/commands.go

package bot

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate, []string, *Command)

type Command struct {
	Name        string
	Description string
	Usage       string
	Cooldown    string
	Hidden      bool
	Handler     string
	Subcommands map[string]*Command
}

type CommandConfig struct {
	Prefix   string
	Commands map[string]*Command
}

var (
	CommandRegistry = make(map[string]*Command)
	HandlerRegistry = make(map[string]CommandHandler)
	CommandHandlers = make(map[string]CommandHandler)
)

func LoadCommands(configPath string, logger *logrus.Logger, handlerRegistry map[string]CommandHandler) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading command config file: %w", err)
	}

	var config CommandConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error unmarshaling command config: %w", err)
	}

	for name, cmd := range config.Commands {
		CommandRegistry[name] = cmd
		logger.Debugf("Registered command: %s (Hidden: %v)", name, cmd.Hidden)

		if handler, exists := handlerRegistry[cmd.Handler]; exists {
			CommandHandlers[name] = handler
			logger.Debugf("Attached handler for command: %s", name)
		} else {
			logger.Warnf("Handler not found for command: %s", name)
		}

		// Handle subcommands
		for subName, subCmd := range cmd.Subcommands {
			fullSubName := name + "." + subName
			logger.Debugf("Processing subcommand: %s for command: %s (Hidden: %v)", subName, name, subCmd.Hidden)

			if subHandler, exists := handlerRegistry[subCmd.Handler]; exists {
				CommandHandlers[fullSubName] = subHandler
				logger.Debugf("Attached handler for subcommand: %s of command: %s", subName, name)
			} else {
				logger.Warnf("Handler not found for subcommand: %s of command: %s", subName, name)
			}
		}
	}

	logger.Infof("Loaded %d commands", len(CommandRegistry))
	return nil
}

func RegisterHandler(name string, handler CommandHandler) {
	HandlerRegistry[name] = handler
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, config *Config) {
	args := strings.Fields(m.Content)
	if len(args) == 0 || !strings.HasPrefix(args[0], config.Discord.CommandPrefix) {
		return
	}

	cmdName := strings.TrimPrefix(args[0], config.Discord.CommandPrefix)
	logrus.Debugf("Handling command: %s", cmdName)

	cmd, exists := CommandRegistry[cmdName]
	if !exists || cmd.Hidden {
		logrus.Debugf("Unknown or hidden command: %s", cmdName)
		SendMessage(s, m.ChannelID, "Unknown command. Use !help to see available commands.")
		return
	}

	if len(args) > 1 && cmd.Subcommands != nil {
		subCmdName := args[1]
		if subCmd, exists := cmd.Subcommands[subCmdName]; exists && !subCmd.Hidden {
			fullSubName := cmdName + "." + subCmdName
			if handler, exists := CommandHandlers[fullSubName]; exists {
				logrus.Debugf("Executing subcommand: %s %s", cmdName, subCmdName)
				handler(s, m, args[2:], subCmd)
				return
			}
			logrus.Warnf("Handler function not found for subcommand: %s %s", cmdName, subCmdName)
		}
	}

	if handler, exists := CommandHandlers[cmdName]; exists {
		logrus.Debugf("Executing command: %s", cmdName)
		handler(s, m, args[1:], cmd)
	} else {
		logrus.Warnf("Handler function not found for command: %s", cmdName)
		SendMessage(s, m.ChannelID, "This command is not implemented.")
	}
}

// func SendMessage(s *discordgo.Session, channelID string, message string) {
//	_, err := s.ChannelMessageSend(channelID, message)
//	if err != nil {
//		fmt.Printf("Error sending message: %v\n", err)
//	}
//}
