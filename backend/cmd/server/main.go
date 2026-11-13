package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"puppy-food/backend/internal/feeding"
	"puppy-food/backend/internal/httpapi"
)

func main() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	insideBackend := filepath.Base(workingDirectory) == "backend"

	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		if insideBackend {
			databasePath = filepath.Join("data", "puppy-food.db")
		} else {
			databasePath = filepath.Join("backend", "data", "puppy-food.db")
		}
	}

	store, err := feeding.Open(databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	service := httpapi.NewService(store)
	var handler http.Handler = service
	dist := os.Getenv("WEB_DIST")
	if dist == "" {
		if insideBackend {
			dist = filepath.Join("..", "frontend", "dist")
		} else {
			dist = filepath.Join("frontend", "dist")
		}
	}
	if _, err := os.Stat(dist); err == nil {
		handler = httpapi.SPA(dist)(handler)
	}

	address := envOr("LISTEN_ADDRESS", ":8080")
	log.Printf("Listening on %s", address)
	if err := http.ListenAndServe(address, handler); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
