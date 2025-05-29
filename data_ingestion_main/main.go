package main

import (
	"log"
	"path/filepath"

	"property_lister/config"
	"property_lister/data_ingestion"
)

func main() {
	config.InitMongo()

	csvPath := filepath.Join("data", "properties.csv")
	if err := data_ingestion.IngestCSV(csvPath); err != nil {
		log.Fatalf("Failed to ingest CSV: %v", err)
	}

	log.Println("Sucessfully migrated data from CSV to MongoDB")
}
