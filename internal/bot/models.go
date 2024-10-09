// File: internal/bot/models.go

package bot

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Database Models

// Term represents a term in the database
type Term struct {
	gorm.Model
	Term        string `gorm:"uniqueIndex;not null"`
	Description string `gorm:"not null"`
}

// Player represents a player in the database
type Player struct {
	DiscordID string `gorm:"primaryKey"`
	PlayerID  string
}

// GiftCodeRedemption represents a gift code redemption in the database
type GiftCodeRedemption struct {
	ID         uint `gorm:"primaryKey"`
	DiscordID  string
	PlayerID   string
	GiftCode   string
	Status     string
	RedeemedAt time.Time
}

// GiftCode represents the structure for gift codes
type GiftCode struct {
	Code        string
	Description string
	Source      string
}

// Scraping Models

// ScrapeSite represents a site configuration for scraping gift codes
type ScrapeSite struct {
	Name     string `mapstructure:"name"`
	URL      string `mapstructure:"url"`
	Selector string `mapstructure:"selector"`
}

// ScrapeResult represents the result of a scrape operation
type ScrapeResult struct {
	SiteName string
	Codes    []GiftCode
	Error    error
}

// Bot Models
type Bot struct {
	Config           *Config
	Session          *discordgo.Session
	DB               *gorm.DB
	logger           *logrus.Logger
	HandlerRegistry  map[string]CommandHandler
	ctx              context.Context
	cancel           context.CancelFunc
	lastCheckedCodes []GiftCode
	scrapeMutex      sync.Mutex
	Code             string
	Description      string
	Source           string
}

// Command represents a bot command and its attributes
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

// Configuration Models

type Config struct {
	Discord struct {
		Token                 string `mapstructure:"token"`
		ClientID              string `mapstructure:"client_id"`
		ClientSecret          string `mapstructure:"client_secret"`
		RoleID                string `mapstructure:"RoleID"`
		RedirectURL           string `mapstructure:"redirect_url"`
		Enabled               bool   `mapstructure:"enabled"`
		CommandPrefix         string `mapstructure:"command_prefix"`
		NotificationChannelID string `mapstructure:"notification_channel_id"`
	} `mapstructure:"discord"`
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Logging struct {
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"logging"`
	Paths struct {
		CommandsConfig string `mapstructure:"commands_config"`
	} `mapstructure:"paths"`
	Database struct {
		VolumeMountPath string `mapstructure:"volumeMountPath"`
		Name            string `mapstructure:"name"`
		Path            string `mapstructure:"path"`
	} `mapstructure:"database"`
	GiftCode struct {
		Salt        string        `mapstructure:"salt"`
		MinLength   int           `mapstructure:"min_length"`
		MaxLength   int           `mapstructure:"max_length"`
		APIEndpoint string        `mapstructure:"api_endpoint"`
		APITimeout  time.Duration `mapstructure:"api_timeout"`
	} `mapstructure:"gift_code"`
	Scrape struct {
		Sites []ScrapeSite `mapstructure:"sites"`
	} `mapstructure:"scrape"`
}
