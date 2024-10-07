// File: internal/bot/database.go
package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB
var dbLogger *logrus.Logger

// Term represents a term in the database
type Term struct {
	gorm.Model
	Term        string `gorm:"uniqueIndex;not null"`
	Description string `gorm:"not null"`
}

// Add these structs at the end of the file
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

func InitDB(config *Config, logger *logrus.Logger) (*gorm.DB, error) {
	dbLogger = logger
	var err error
	db, err = gorm.Open(sqlite.Open(config.Database.Path), &gorm.Config{
		Logger: CustomGormLogger{Logger: logger},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	return db, nil
}

//func RunMigrations(db *gorm.DB) error {
//	return db.AutoMigrate(&Term{}, &Player{}, &GiftCodeRedemption{})
//}

// Term-related functions

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

// Player-related functions

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

func RemovePlayerID(discordID string) error {
	result := db.Model(&Player{}).Where("discord_id = ?", discordID).Update("player_id", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no player found with the given Discord ID")
	}
	return nil
}

// Add this function
func RemovePlayer(discordID string) error {
	result := db.Where("discord_id = ?", discordID).Delete(&Player{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no player found with the given Discord ID")
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

func GetPlayerID(discordID string) (string, error) {
	var player Player
	result := db.Where("discord_id = ?", discordID).First(&player)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("no player found with Discord ID: %s", discordID)
		}
		return "", result.Error
	}
	return player.PlayerID, nil
}

func GetAllPlayerIDs() (map[string]string, error) {
	var players []Player
	result := db.Find(&players)
	if result.Error != nil {
		return nil, result.Error
	}

	playerIDs := make(map[string]string)
	for _, player := range players {
		playerIDs[player.DiscordID] = player.PlayerID
	}
	return playerIDs, nil
}

// GiftCode-related functions

func RecordGiftCodeRedemption(discordID, playerID, giftCode, status string) error {
	redemption := GiftCodeRedemption{
		DiscordID:  discordID,
		PlayerID:   playerID,
		GiftCode:   giftCode,
		Status:     status,
		RedeemedAt: time.Now(),
	}
	return db.Create(&redemption).Error
}

// TODO: Implement these functions if needed
func GetAllGiftCodeRedemptionsPaginated(page, itemsPerPage int) ([]GiftCodeRedemption, error) {
	var redemptions []GiftCodeRedemption
	offset := (page - 1) * itemsPerPage
	result := db.Order("redeemed_at desc").Offset(offset).Limit(itemsPerPage).Find(&redemptions)
	return redemptions, result.Error
}

func GetUserGiftCodeRedemptionsPaginated(discordID string, page, itemsPerPage int) ([]GiftCodeRedemption, error) {
	var redemptions []GiftCodeRedemption
	offset := (page - 1) * itemsPerPage
	result := db.Where("discord_id = ?", discordID).Order("redeemed_at desc").Offset(offset).Limit(itemsPerPage).Find(&redemptions)
	return redemptions, result.Error
}
