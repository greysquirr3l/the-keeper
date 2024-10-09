// utility.go
package bot

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
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
