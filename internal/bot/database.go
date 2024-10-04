// database.go
package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB
var dbLogger *logrus.Logger

// Player represents a player in the database
type Player struct {
	gorm.Model
	DiscordID     string `gorm:"uniqueIndex;not null"`
	PlayerID      string `gorm:"uniqueIndex;not null"`
	GiftsRedeemed string
}

// Term represents a term in the database
type Term struct {
	gorm.Model
	Term        string `gorm:"uniqueIndex;not null"`
	Description string `gorm:"not null"`
}

// CustomGormLogger adapts logrus to implement gorm's logger interface
type CustomGormLogger struct {
	Logger *logrus.Logger
}

func (l CustomGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l CustomGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Infof(msg, data...)
}

func (l CustomGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warnf(msg, data...)
}

func (l CustomGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Errorf(msg, data...)
}

func (l CustomGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		l.Logger.Errorf("TRACE: %s, %d rows, %v, %v", sql, rows, elapsed, err)
	} else {
		l.Logger.Infof("TRACE: %s, %d rows, %v", sql, rows, elapsed)
	}
}

func InitDB(config *Config, logger *logrus.Logger) error {
	dbLogger = logger
	var err error
	db, err = gorm.Open(sqlite.Open(config.Database.Path), &gorm.Config{
		Logger: CustomGormLogger{Logger: logger},
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func AddTerm(term, description string) error {
	return db.Create(&Term{Term: term, Description: description}).Error
}

func EditTerm(term, newDescription string) error {
	result := db.Model(&Term{}).Where("term = ?", term).Update("description", newDescription)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("term not found")
	}
	return nil
}

func RemoveTerm(term string) error {
	result := db.Where("term = ?", term).Delete(&Term{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("term not found")
	}
	return nil
}

func ListTerms() ([]Term, error) {
	var terms []Term
	if err := db.Order("term").Find(&terms).Error; err != nil {
		return nil, err
	}
	return terms, nil
}

func GetTerm(term string) (*Term, error) {
	var t Term
	if err := db.Where("term = ?", term).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func AddPlayer(discordID, playerID string) error {
	return db.Create(&Player{DiscordID: discordID, PlayerID: playerID}).Error
}

func EditPlayerID(discordID, newPlayerID string) error {
	result := db.Model(&Player{}).Where("discord_id = ?", discordID).Update("player_id", newPlayerID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("player not found")
	}
	return nil
}

func ListPlayers() ([]Player, error) {
	var players []Player
	if err := db.Order("discord_id").Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}
