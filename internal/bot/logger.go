// internal/bot/logger.go
package bot

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Log is the global logger used throughout the app
var Log = logrus.New()

// InitializeLogger sets up logrus formatting and log level based on the configuration
func InitializeLogger(config *Config) {
	// Set the output to stdout
	Log.SetOutput(os.Stdout)

	// Set the log level based on the configuration
	switch config.Logging.LogLevel {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel) // Default to Info level if no valid log level is provided
	}

	// Set the log format to TextFormatter with full timestamps
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	Log.Infof("Logger initialized with level: %s", config.Logging.LogLevel)
}
