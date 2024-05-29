package grf

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupDatabase() (*mongo.Database, func(), error) {
	connectionString, dbname := GetDBDetails()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))

	disconnect := func() {
		cancel()
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}
	db := client.Database(dbname)

	return db, disconnect, err
}

func GetDBDetails() (uri, dbname string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	uri = os.Getenv("DATABASE_URI")
	dbname = os.Getenv("DATABASE_NAME")
	return
}
