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

func RegisterHandler(name string, handler CommandHandler) {
	HandlerRegistry[name] = handler
}

func LoadCommands(configPath string) error {
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
		if cmd.Handler != "" {
			handler, exists := HandlerRegistry[cmd.Handler]
			if !exists {
				logrus.Warnf("Handler %s not found for command %s, using placeholder", cmd.Handler, name)
				cmd.HandlerFunc = HandlerRegistry["placeholderHandler"]
			} else {
				cmd.HandlerFunc = handler
			}
		} else {
			logrus.Warnf("No handler specified for command %s, using placeholder", name)
			cmd.HandlerFunc = HandlerRegistry["placeholderHandler"]
		}

		for subName, subCmd := range cmd.Subcommands {
			if subCmd.Handler != "" {
				handler, exists := HandlerRegistry[subCmd.Handler]
				if !exists {
					logrus.Warnf("Handler %s not found for subcommand %s.%s, using placeholder", subCmd.Handler, name, subName)
					subCmd.HandlerFunc = HandlerRegistry["placeholderHandler"]
				} else {
					subCmd.HandlerFunc = handler
				}
			} else {
				logrus.Warnf("No handler specified for subcommand %s.%s, using placeholder", name, subName)
				subCmd.HandlerFunc = HandlerRegistry["placeholderHandler"]
			}
		}

		CommandRegistry[name] = cmd
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
		if subExists {
			subCmd.HandlerFunc(s, m, args[2:], subCmd)
			return
		}
	}

	cmd.HandlerFunc(s, m, args[1:], cmd)
}
