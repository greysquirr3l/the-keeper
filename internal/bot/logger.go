package bot

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Initialize the global logger
var Log = logrus.New()

// InitializeLogger sets up logrus formatting and output settings
func InitializeLogger(config *Config) {
	// Set the output to stdout
	Log.SetOutput(os.Stdout)

	// Set the log level from config (can be "info", "debug", etc.)
	switch config.Logging.LogLevel {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	// Set the format to JSON (optional: change to TextFormatter if preferred)
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
