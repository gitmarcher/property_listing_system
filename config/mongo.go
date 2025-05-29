package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongo() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error while loading .env file", err)
	}

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	err = mgm.SetDefaultConfig(nil, dbName, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	log.Println("Successfully Connected to MongoDB")
}
