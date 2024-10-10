// migrations.go
package bot

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func runMigrations(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202310071200", // Static ID for creating the terms table
			Migrate: func(tx *gorm.DB) error {
				// Create the terms table
				type Term struct {
					gorm.Model
					Term        string `gorm:"uniqueIndex;not null"`
					Description string `gorm:"not null"`
				}
				return tx.AutoMigrate(&Term{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("terms")
			},
		},
		{
			ID: "202310071300", // Static ID for creating the players table
			Migrate: func(tx *gorm.DB) error {
				// Create the players table
				type Player struct {
					DiscordID string `gorm:"primaryKey"`
					PlayerID  string
				}
				return tx.AutoMigrate(&Player{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("players")
			},
		},
		{
			ID: "202310071400", // Static ID for creating the gift_code_redemptions table
			Migrate: func(tx *gorm.DB) error {
				// Create the gift_code_redemptions table
				type GiftCodeRedemption struct {
					ID         uint `gorm:"primaryKey"`
					DiscordID  string
					PlayerID   string
					GiftCode   string
					Status     string
					RedeemedAt time.Time
				}
				return tx.AutoMigrate(&GiftCodeRedemption{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("gift_code_redemptions")
			},
		},
	})

	// Run the migrations
	if err := m.Migrate(); err != nil {
		dbLogger.WithError(err).Error("Could not migrate")
		return err
	}

	dbLogger.Info("Migration ran successfully")
	return nil
}

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&Term{}, &Player{}, &GiftCodeRedemption{})
}
