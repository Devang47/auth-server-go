package main

import (
	"auth-server-go/internal/database"
	"auth-server-go/internal/server"
	"fmt"
	"log"
)

func main() {
	db, err := database.SetupDatabase()
	if err != nil {
		panic(fmt.Sprintf("cannot connect to database: %s", err))
	}
	server.SetupGoogleAuth()
	server := server.NewServer(db)

	log.Print("Server started on port ", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
