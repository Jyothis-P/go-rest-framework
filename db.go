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

// Connects to mongodb using details fetched form the env variables.
func SetupDatabase() (*mongo.Database, func(), error) {
	connectionString, dbname := GetDBDetails(false)
	return getDBConnection(connectionString, dbname)
}

func SetupTestDatabase() (*mongo.Database, func(), error) {
	connectionString, dbname := GetDBDetails(true)
	return getDBConnection(connectionString, dbname)
}

func getDBConnection(connectionString, dbname string) (*mongo.Database, func(), error) {
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

// Loads the env file and gets the connection string and db name.
// Takes a boolean value to specificy whether to use test db or not.
func GetDBDetails(test bool) (uri, dbname string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	uri = os.Getenv("DATABASE_URI")
	if test {
		dbname = "test_db"
	} else {
		dbname = os.Getenv("DATABASE_NAME")
	}
	return
}
