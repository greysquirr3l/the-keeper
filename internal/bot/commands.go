// File: internal/bot/commands.go

package bot

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bwmarrin/discordgo"
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

	CommandRegistry = config.Commands
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
		if subExists && subCmd.Handler != "" {
			if handler, ok := HandlerRegistry[subCmd.Handler]; ok {
				handler(s, m, args[2:], subCmd)
				return
			}
		}
	}

	if cmd.Handler != "" {
		if handler, ok := HandlerRegistry[cmd.Handler]; ok {
			handler(s, m, args[1:], cmd)
		} else {
			SendMessage(s, m.ChannelID, "This command is not implemented.")
		}
	} else {
		SendMessage(s, m.ChannelID, "This command is not implemented.")
	}
}

// func SendMessage(s *discordgo.Session, channelID string, message string) {
//	_, err := s.ChannelMessageSend(channelID, message)
//	if err != nil {
//		fmt.Printf("Error sending message: %v\n", err)
//	}
//}
