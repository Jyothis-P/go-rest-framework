package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	grf "github.com/Jyothis-P/go-rest-framework"
	"github.com/gorilla/mux"
)

// Create a model
// Make sure to give json and bson structs as necessary.
type Todo struct {
	Title     string `json:"title" bson:"title"`
	Completed bool   `json:"completed" bson:"completed"`
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
	grf.RegisterCRUDRoutes[Todo]("/todo", r, &appContext)

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
