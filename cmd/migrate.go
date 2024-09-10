package main

import (
	"auth-server-go/internal/database"
	models "auth-server-go/internal/models"
	"fmt"
	"log"
)

func main() {
	// Set up database
	db, err := database.SetupDatabase()
	if err != nil {
		log.Fatalf("Could not setup database: %v", err)
	}

	fmt.Println("Migrating schemas...")
	// Migrate the schemas
	for _, model := range models.DatabaseModels {
		db.AutoMigrate(model)
	}

	fmt.Println("Finished Migrating Schemas")
}
