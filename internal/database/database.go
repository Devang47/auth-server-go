package database

import (
	"auth-server-go/internal/models"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type service struct {
	db *gorm.DB
}

var (
	dbInstance *service
)

func SetupDatabase() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_DATABASE")

	var connectionString = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=require",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
	)

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sqlDB: %v", err)
	}

	sqlDB.SetConnMaxLifetime(time.Minute * 5)
	sqlDB.SetConnMaxIdleTime(time.Minute * 5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	return db, nil
}

func New() *service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := SetupDatabase()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) MigrateSchema() {
	for _, model := range models.DatabaseModels {
		s.db.AutoMigrate(model)
	}
}
