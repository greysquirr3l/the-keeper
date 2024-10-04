package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// parseCommand extracts the main command and subcommand from the message content.
func parseCommand(content string) (string, string) {
	// Split the message content by spaces to extract the command and subcommand
	parts := strings.Fields(content)

	// The first part is the command name (e.g., "id")
	if len(parts) > 0 {
		cmdName := parts[0]

		// If there's a subcommand, it would be the second part (e.g., "add" in "!id add 1234")
		if len(parts) > 1 {
			subCmdName := parts[1]
			return cmdName, subCmdName
		}

		// If there's no subcommand, just return the command name
		return cmdName, ""
	}

	// If no command was found, return empty strings
	return "", ""
}

// HandleCommand processes incoming commands
func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, commandConfig *CommandConfig) {
	// Extract the command prefix from the configuration
	prefix := commandConfig.Prefix
	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	// Parse the command and subcommand from the message
	cmdName, subCmdName := parseCommand(m.Content[len(prefix):])

	// Find the command in the configuration
	cmd, exists := commandConfig.Commands[cmdName]
	if !exists {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Command '%s' not found.", cmdName))
		Log.Warnf("Command '%s' not found", cmdName)
		return
	}

	var cooldown time.Duration
	var err error

	// Check if a subcommand is provided
	if subCmdName != "" {
		subCmd, subExists := cmd.Subcommands[subCmdName]
		if !subExists {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Subcommand '%s' not found under command '%s'.", subCmdName, cmdName))
			Log.Warnf("Subcommand '%s' not found under command '%s'", subCmdName, cmdName)
			return
		}

		// Parse cooldown for the subcommand
		cooldown, err = time.ParseDuration(subCmd.Cooldown)
		if err != nil {
			Log.Errorf("Invalid cooldown format for subcommand '%s': %v", subCmdName, err)
			return
		}

		// Handle subcommand execution and cooldown
		if !CheckCooldown(m.Author.ID, cmdName+"_"+subCmdName, int(cooldown.Seconds())) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You must wait before using the '%s' subcommand again.", subCmdName))
			return
		}

		executeSubCommand(s, m, subCmd)
		SetCooldown(m.Author.ID, cmdName+"_"+subCmdName)
		Log.Infof("Subcommand '%s' executed by user %s", subCmdName, m.Author.Username)
	} else {
		// Parse cooldown for the main command
		cooldown, err = time.ParseDuration(cmd.Cooldown)
		if err != nil {
			Log.Errorf("Invalid cooldown format for command '%s': %v", cmdName, err)
			return
		}

		// Handle main command execution and cooldown
		if !CheckCooldown(m.Author.ID, cmdName, int(cooldown.Seconds())) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You must wait before using the '%s' command again.", cmdName))
			return
		}
		executeCommand(s, m, cmd)
		SetCooldown(m.Author.ID, cmdName)
		Log.Infof("Command '%s' executed by user %s", cmdName, m.Author.Username)
	}
}

// executeCommand processes and executes a main command dynamically from the YAML config.
func executeCommand(s *discordgo.Session, m *discordgo.MessageCreate, cmd Command) {
	// Log the command execution
	Log.Infof("Executing command '%s' by user %s", cmd.Description, m.Author.Username)

	// Check if the command is hidden and skip execution if it is
	if cmd.Hidden {
		Log.Infof("Command '%s' is hidden and won't be executed.", cmd.Description)
		return
	}

	// Example logic to show command execution (replace this with your specific logic)
	message := fmt.Sprintf("Command '%s' executed with usage: %s", cmd.Description, cmd.Usage)
	s.ChannelMessageSend(m.ChannelID, message)

	// You can add additional logic for each command here
}

// executeSubCommand dynamically processes and executes a subcommand based on the configuration.
func executeSubCommand(s *discordgo.Session, m *discordgo.MessageCreate, subCmd Subcommand) {
	// Log the subcommand execution
	Log.Infof("Executing subcommand '%s' by user %s", subCmd.Description, m.Author.Username)

	// Check if the subcommand is hidden and skip execution
	if subCmd.Hidden {
		Log.Infof("Subcommand '%s' is hidden and won't be executed.", subCmd.Description)
		return
	}

	// Example logic to show subcommand execution. You can replace this with specific logic.
	message := fmt.Sprintf("Subcommand '%s' executed with usage: %s", subCmd.Description, subCmd.Usage)
	s.ChannelMessageSend(m.ChannelID, message)
}
