// File: internal/bot/database.go

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

// CustomGormLogger adapts logrus to implement gorm's logger interface
type CustomGormLogger struct {
	Logger   *logrus.Logger
	logLevel gormlogger.LogLevel
}

func (l CustomGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := l
	newLogger.logLevel = level
	return newLogger
}

func (l CustomGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.Logger.Infof("[GORM INFO] "+msg, data...)
	}
}

func (l CustomGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.Logger.Warnf("[GORM WARN] "+msg, data...)
	}
}

func (l CustomGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.Logger.Errorf("[GORM ERROR] "+msg, data...)
	}
}

func (l CustomGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel >= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.logLevel >= gormlogger.Error:
		l.Logger.Errorf("[GORM TRACE ERROR] %s, %d rows affected, %v, %v", sql, rows, elapsed, err)
	case elapsed > 200*time.Millisecond && l.logLevel >= gormlogger.Warn:
		l.Logger.Warnf("[GORM SLOW QUERY] %s, %d rows affected, %v", sql, rows, elapsed)
	case l.logLevel >= gormlogger.Info:
		l.Logger.Infof("[GORM TRACE] %s, %d rows affected, %v", sql, rows, elapsed)
	}
}

func InitDB(config *Config, logger *logrus.Logger) (*gorm.DB, error) {
	dbLogger = logger
	var err error

	// Open the database connection
	db, err = gorm.Open(sqlite.Open(config.Database.Path), &gorm.Config{
		Logger: CustomGormLogger{
			Logger:   logger,
			logLevel: gormlogger.Info, // Set default log level here (Silent, Error, Warn, Info)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Set database to use Write-Ahead Logging (WAL) mode for better durability
	if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Run the migrations using gormigrate
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
