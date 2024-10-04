package bot

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Global cooldown tracker to ensure users respect command cooldowns
var cooldowns = make(map[string]time.Time)
var mu sync.Mutex

// CheckCooldown checks if a user can use a command or if they are still in the cooldown period.
func CheckCooldown(userID, command string, cooldownSeconds int) bool {
	mu.Lock()
	defer mu.Unlock()

	// Get the current time and the time when the command can next be used
	now := time.Now()
	cooldownKey := userID + "_" + command
	nextAllowedTime, exists := cooldowns[cooldownKey]

	// If there's no record of the cooldown, allow the command
	if !exists || now.After(nextAllowedTime) {
		return true
	}

	Log.WithFields(logrus.Fields{
		"user_id":      userID,
		"command":      command,
		"next_allowed": nextAllowedTime,
	}).Info("Cooldown active, user must wait to use the command again")

	// If the current time is before the next allowed time, return false
	return false
}

// SetCooldown sets the cooldown for a user and a specific command.
func SetCooldown(userID, command string) {
	mu.Lock()
	defer mu.Unlock()

	// Create the cooldown key and set the cooldown expiration time
	cooldownKey := userID + "_" + command
	cooldownDuration := time.Duration(5) * time.Second // Default 5 seconds for simplicity, adjust as needed

	cooldowns[cooldownKey] = time.Now().Add(cooldownDuration)

	Log.WithFields(logrus.Fields{
		"user_id":  userID,
		"command":  command,
		"cooldown": cooldownDuration,
	}).Info("Command cooldown set")
}

// ResetCooldown manually resets the cooldown for a user and a specific command.
func ResetCooldown(userID, command string) {
	mu.Lock()
	defer mu.Unlock()

	cooldownKey := userID + "_" + command
	delete(cooldowns, cooldownKey)

	Log.WithFields(logrus.Fields{
		"user_id": userID,
		"command": command,
	}).Info("Command cooldown reset")
}

// LogError logs an error message.
func LogError(err error, message string) {
	Log.WithError(err).Error(message)
}
