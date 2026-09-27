package db

import (
	"GoPower/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		dsn = "file::memory:?cache=shared"
	}

	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	err = database.AutoMigrate(
		&models.PowerStation{},
		&models.ConsumerMeter{},
		&models.DispatchRecord{},
	)
	if err != nil {
		return nil, err
	}

	return database, nil
}
