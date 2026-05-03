package repository

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection and auto-migrates models.
func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to the database")

	// Auto-migrate models
	if err := db.AutoMigrate(&Pair{}, &PairCategory{}); err != nil {
		log.Printf("Failed to auto migrate database: %v\n", err)
		return nil, err
	}

	return db, nil
}
