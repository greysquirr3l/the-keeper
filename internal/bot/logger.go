package bot

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Initialize the global logger
var log = logrus.New()

// InitializeLogger sets up logrus formatting and output settings
func InitializeLogger() {
	// Set the output to stdout
	log.SetOutput(os.Stdout)

	// Set the log level (can adjust to Debug, Warn, Error, etc.)
	log.SetLevel(logrus.InfoLevel)

	// Set the format to JSON (optional: change to TextFormatter if preferred)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
