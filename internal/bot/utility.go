// utility.go
package bot

import (
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
)

func init() {
	cooldownCache = cache.New(5*time.Minute, 10*time.Minute)
}

func SetUtilLogger(logger *logrus.Logger) {
	utilLogger = logger
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
