package database

import (
	"commitcaster/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect подключается к PostgreSQL и выполняет миграции
func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	// Автоматические миграции
	if err := AutoMigrate(); err != nil {
		return err
	}

	return nil
}

// AutoMigrate выполняет автоматические миграции
func AutoMigrate() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&models.User{},
		&models.UserSettings{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("✅ Database migrations completed")
	return nil
}

// GetDB возвращает экземпляр БД
func GetDB() *gorm.DB {
	return DB
}
