package grf

import (
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

// Application Context struct. This is made available to your handler functions as a parameter.
// Handy for keeping values like database connections.
type Ctx struct {
	DB *mongo.Database
}

// An adapter for handler functions with an added app context passed in.
// Implements http.Handler.
type H struct {
	*Ctx
	Fn func(*Ctx, http.ResponseWriter, *http.Request)
}

func (appHandler H) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	appHandler.Fn(appHandler.Ctx, w, r)
}
