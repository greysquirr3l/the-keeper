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

type Command struct {
	Name        string
	Description string
	Usage       string
	Cooldown    string
	Handler     string
	Hidden      bool
	Subcommands map[string]*Command
	HandlerFunc func(*discordgo.Session, *discordgo.MessageCreate, []string, *Command)
}

var CommandRegistry map[string]*Command

func LoadCommands(configPath string, logger *logrus.Logger, handlerRegistry map[string]CommandHandler) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading command config file: %w", err)
	}
	var config struct {
		Prefix   string
		Commands map[string]*Command
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error parsing command config: %w", err)
	}
	CommandRegistry = config.Commands
	logger.Infof("Loaded %d commands from config", len(CommandRegistry))

	for name, cmd := range CommandRegistry {
		cmd.Name = name
		if handler, ok := handlerRegistry[cmd.Handler]; ok {
			cmd.HandlerFunc = handler
			logger.Infof("Handler '%s' associated with command '%s'", cmd.Handler, name)
		} else {
			logger.Warnf("Handler '%s' not found for command '%s'", cmd.Handler, name)
		}

		for subName, subCmd := range cmd.Subcommands {
			subCmd.Name = subName
			if handler, ok := handlerRegistry[subCmd.Handler]; ok {
				subCmd.HandlerFunc = handler
				logger.Infof("Handler '%s' associated with subcommand '%s' of '%s'", subCmd.Handler, subName, name)
			} else {
				logger.Warnf("Handler '%s' not found for subcommand '%s' of '%s'", subCmd.Handler, subName, name)
			}
		}
	}

	logger.Info("Command loading completed")
	return nil
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, config *Config) {
	content := strings.TrimPrefix(m.Content, config.Discord.CommandPrefix)
	args := strings.Fields(content)

	if len(args) == 0 {
		return
	}

	cmdName := strings.ToLower(args[0])
	cmd, exists := CommandRegistry[cmdName]

	if !exists {
		return
	}

	if !CheckCooldown(m.Author.ID, cmdName, cmd.Cooldown) {
		return
	}

	if cmd.HandlerFunc != nil {
		cmd.HandlerFunc(s, m, args[1:], cmd)
	} else {
		SendMessage(s, m.ChannelID, fmt.Sprintf("Command '%s' is not implemented yet.", cmdName))
	}
}
