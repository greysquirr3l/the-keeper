// utility.go
package bot

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	cooldownCache *cache.Cache
	cacheMutex    sync.Mutex
	utilLogger    *logrus.Logger
	logger        *logrus.Logger
	loggerOnce    sync.Once
)

func init() {
	cooldownCache = cache.New(5*time.Minute, 10*time.Minute)
}

func SetUtilLogger(logger *logrus.Logger) {
	utilLogger = logger
}

// // TODO: is GetLogger correct?
func GetLogger() *logrus.Entry {
	loggerOnce.Do(func() {
		logger = logrus.New()
		logger.SetLevel(logrus.DebugLevel)
	})
	return logger.WithField("service", "the-keeper")
}

func ParseArguments(input string) []string {
	return strings.Fields(input)
}

func CheckCooldown(userID, command, cooldownStr string) bool {
	if cooldownStr == "" {
		return true
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	key := userID + ":" + command
	if _, found := cooldownCache.Get(key); found {
		return false
	}

	duration, err := time.ParseDuration(cooldownStr)
	if err != nil {
		utilLogger.Errorf("Invalid cooldown duration: %v", err)
		return true
	}

	cooldownCache.Set(key, true, duration)
	return true
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

// NormalizeInput trims spaces and converts input to lowercase for consistent command handling.
func NormalizeInput(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func (b *Bot) GetPlayerID(discordID string) (string, error) {
	var player Player
	err := b.DB.Where("discord_id = ?", discordID).First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("no player found with Discord ID: %s", discordID)
		}
		return "", err
	}
	return player.PlayerID, nil
}

func (b *Bot) GetAllPlayerIDs() (map[string]string, error) {
	var players []Player
	result := b.DB.Find(&players)
	if result.Error != nil {
		return nil, result.Error
	}

	playerIDs := make(map[string]string)
	for _, player := range players {
		playerIDs[player.DiscordID] = player.PlayerID
	}
	return playerIDs, nil
}

// ListPlayers lists all players in the database.
func (b *Bot) ListPlayers() ([]Player, error) {
	var players []Player
	result := b.DB.Order("discord_id").Find(&players)
	if result.Error != nil {
		return nil, result.Error
	}
	return players, nil
}
