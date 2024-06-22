package grf

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Generic function to add objects to the database.
// models.Object is stored in the objects collection.
// Automatically adds the record to the collection with a plural, lowercase name.
func Create[K interface{}](database *mongo.Database, object K) (*mongo.InsertOneResult, error) {
	collection, ctx, cancel := getCollectionAndContext(database, object)
	defer cancel()
	res, err := collection.InsertOne(ctx, object)
	if err != nil {
		log.Println("Error adding object to database.", err)
		return nil, err
	}
	log.Println("Inserted record to " + collection.Name() + " collection.")
	return res, err
}

// Reads all the objects of the given type.
func Read[K any](database *mongo.Database, objects *[]K) error {
	collection, ctx, cancel := getCollectionAndContext(database, objects)
	defer cancel()

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Println("error retrieving all objects of "+collection.Name(), err)
		return err
	}
	err = cur.All(ctx, objects)
	if err != nil {
		log.Println("error getting data from cursor "+collection.Name(), err)
		return err
	}
	return nil
}

func ReadOne[K any](database *mongo.Database, object *K, id string) error {
	collection, ctx, cancel := getCollectionAndContext(database, object)
	defer cancel()

	// Converting the id from the hex string to the ObjectID format that mongo use
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error converting id to ObjectId:", err)
		return err
	}
	filter := bson.D{{Key: "_id", Value: objectID}}
	log.Println("Filter: ", filter)
	err = collection.FindOne(ctx, filter).Decode(&object)
	log.Println("Object: ", object)
	if err != nil {
		log.Println("Error finding the one", err)
		return err
	}
	return nil
}

func ReplaceOne[K any](database *mongo.Database, object *K, id string) error {
	collection, ctx, cancel := getCollectionAndContext(database, object)
	defer cancel()

	// Converting the id from the hex string to the ObjectID format that mongo use
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error converting id to ObjectId:", err)
		return err
	}
	filter := bson.D{{Key: "_id", Value: objectID}}
	res, err := collection.ReplaceOne(ctx, filter, *object)
	if err != nil {
		log.Println("Error replacing object:", err)
		return err
	}
	log.Println("Replaced object.", res.ModifiedCount)
	return nil
}

// about as dumb as it gets. Works for atomics. Wouldn't recommend for anything with dependencies.
func Delete[K any](database *mongo.Database, id string) error {
	collection, ctx, cancel := getCollectionAndContext(database, *new(K))
	defer cancel()

	// Converting the group id from the hex string to the ObjectID format that mongo use
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error converting id to ObjectId:", err)
		return err
	}
	filter := bson.D{{Key: "_id", Value: objectID}}
	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println("Error deleting object:", err)
		return err
	}
	log.Println("Deleted object.", res.DeletedCount)
	return nil
}

func getPlural(noun string) string {
	// Model type name could be a variation of the following.
	// package.model, *package.model, []package.model, *[]package.model
	// We're extracting the model name and returning its plural.
	// "*models.Todo" becomes "todos" || "*[]models.Box" becomes "boxes"
	splits := strings.Split(strings.TrimPrefix(strings.TrimPrefix(noun, "*"), "[]"), ".")
	noun = strings.ToLower(splits[len(splits)-1])
	ends := []string{"s", "sh", "ch", "x", "z"}
	for _, end := range ends {
		if strings.HasSuffix(noun, end) {
			return noun + "es"
		}
	}
	return noun + "s"
}

func getCollection[K any](database *mongo.Database) *mongo.Collection {
	object := new(K)
	collectionName := getPlural(fmt.Sprintf("%T", object))
	collection := database.Collection(collectionName)
	return collection
}

func getCollectionAndContext[K any](database *mongo.Database, object K) (*mongo.Collection, context.Context, context.CancelFunc) {
	collectionName := getPlural(fmt.Sprintf("%T", object))
	collection := database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	return collection, ctx, cancel
}
