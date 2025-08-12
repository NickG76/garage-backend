//main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nickg76/garage-backend/internal/handlers"
	"github.com/nickg76/garage-backend/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := handlers.NewServer()
	mux := server.Routes(srv)

	log.Printf("Listening on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
