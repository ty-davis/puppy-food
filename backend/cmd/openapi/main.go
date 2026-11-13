package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"puppy-food/backend/internal/feeding"
	"puppy-food/backend/internal/httpapi"
)

func main() {
	store, err := feeding.Open(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	service := httpapi.NewService(store)
	schema, err := json.MarshalIndent(service.OpenAPISchema(), "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join("openapi.json")
	if err := os.WriteFile(path, append(schema, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote %s", path)
}
