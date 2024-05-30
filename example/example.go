package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	grf "github.com/Jyothis-P/go-rest-framework"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create a model
// Make sure to give json and bson structs as necessary.
type Todo struct {
	Id        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Completed bool               `json:"completed" bson:"completed"`
}

func (todo *Todo) markCompleted(completed bool) {
	todo.Completed = completed
}

func main() {

	// Setup database.
	db, cancel, err := grf.SetupDatabase()
	defer cancel()
	if err != nil {
		log.Println("Error setting up the database.", err)
		return
	}

	// Create App context.
	appContext := grf.Ctx{
		DB: db,
	}

	// Create a router.
	r := mux.NewRouter().StrictSlash(true)

	// Register routes for the model.
	todoRouter := grf.RegisterCRUDRoutes[Todo]("/todo", r, &appContext)

	todoRouter.Handle("/{id}/customDeleteTodo", grf.H{Ctx: &appContext, Fn: customDeleteHandler}).Methods("DELETE")
	todoRouter.Handle("/{id}/markComplete", grf.H{Ctx: &appContext, Fn: markComplete}).Methods("PUT")

	// Set up server.
	const PORT string = "8001"
	srv := &http.Server{
		Addr:         "0.0.0.0:" + PORT,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// Start the server.

	go func() {
		log.Println("Starting web server on " + PORT)
		if err := srv.ListenAndServe(); err != nil {
			log.Println("Error with the web server.")
			log.Println(err)
		}
	}()

	// Graceful shutdown of the server in case of os interrupts.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Shutting down.")
	os.Exit(0)
}

// Customer Handler function using the generic Delete service.
func customDeleteHandler(appCtx *grf.Ctx, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// You can use any of the generic service functions or your own custom service.
	err := grf.Delete[Todo](appCtx.DB, vars["id"])
	if err != nil {
		http.Error(w, "Error deleting TODO.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Group deleted.")
}

// Custom Handler with a custom service making use of the db connection passed from context.
func markComplete(appCtx *grf.Ctx, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := finisherService(appCtx.DB, vars["id"])
	if err != nil {
		http.Error(w, "Error completing TODO.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Marked complete!")
}

func finisherService(db *mongo.Database, id string) error {
	collection := db.Collection("todos")

	// Converting the group id from the hex string to the ObjectID format that mongo use
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error converting id to ObjectId:", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	filter := bson.D{{Key: "_id", Value: objectID}}
	update := bson.M{"$set": bson.M{"completed": true}}
	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil || res.ModifiedCount < 1 {
		log.Println("Error Marking complete:", err)
		log.Println("Matched Count: ", res.MatchedCount)
		log.Println("Modified Count: ", res.ModifiedCount)
		return err
	}
	return nil
}
