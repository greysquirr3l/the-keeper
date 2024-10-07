// migrations.go
package bot

import (
	"gorm.io/gorm"
)

// func RunMigrations(db *gorm.DB) error {
//	return db.AutoMigrate(&Term{}, &Player{})
// }

func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&Term{}, &Player{}, &GiftCodeRedemption{})
}
