package grf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Function to register the basic CRUD routes given a model.
// User can register individual routes from the generic handlers.
// Or they can use this function to generate teh default REST endpoints.
// Problem comes when it is mongodb which does not support CASCADE delete out of the box and you need a better logic for deleting.
// Maybe add an optional callback function?
// Putting a pin on it. [This can be good for atomics]
// [For objects with more complex dependencies, use the handlers you need and create the rest yourself]
func RegisterCRUDRoutes[T any](pathPrefix string, r *mux.Router, ctx *Ctx) *mux.Router {
	subRouter := r.PathPrefix(pathPrefix).Subrouter()
	subRouter.Handle("/", H{Ctx: ctx, Fn: GetAllHandler[T]}).Methods("GET")
	subRouter.Handle("/{id}", H{Ctx: ctx, Fn: GetHandler[T]}).Methods("GET")
	subRouter.Handle("/{id}", H{Ctx: ctx, Fn: ReplaceHandler[T]}).Methods("PUT")
	subRouter.Handle("/", H{Ctx: ctx, Fn: CreateHandler[T]}).Methods("POST")
	subRouter.Handle("/{id}", H{Ctx: ctx, Fn: DeleteHandler[T]}).Methods("DELETE")
	return subRouter
}

// Adds Read and ReadOne routes for type T to the router.
// GET /
// GET /{id}
func AddReadRoutes[T any](r *mux.Router, ctx *Ctx) {
	r.Handle("/", H{Ctx: ctx, Fn: GetAllHandler[T]}).Methods("GET")
	r.Handle("/{id}", H{Ctx: ctx, Fn: GetHandler[T]}).Methods("GET")
}

// Adds Delete route for type T to the router.
// DELETE /{id}
func AddDeleteRoute[T any](r *mux.Router, ctx *Ctx) {
	r.Handle("/{id}", H{Ctx: ctx, Fn: DeleteHandler[T]}).Methods("DELETE")
}

// Adds Create route for type T to the router.
// POST /
// body must containt the object as defined by the model and its struct tags.
func AddCreateRoute[T any](r *mux.Router, ctx *Ctx) {
	r.Handle("/", H{Ctx: ctx, Fn: CreateHandler[T]}).Methods("POST")
}

// Adds Replace route for type T to the router.
// PUT /{id}
// body must contain the entire object with the required changes.
// if any field is not supplied(except _id), it will be reset to its nil value.
func AddReplaceRoute[T any](r *mux.Router, ctx *Ctx) {
	r.Handle("/{id}", H{Ctx: ctx, Fn: ReplaceHandler[T]}).Methods("PUT")
}

func GetHandler[K any](ctx *Ctx, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var object K
	err := ReadOne(ctx.DB, &object, vars["id"])
	if err != nil {
		log.Print("Error retrieving object.")
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(object)
	if err != nil {
		log.Print("Error marshalling.")
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func GetAllHandler[K any](ctx *Ctx, w http.ResponseWriter, r *http.Request) {
	var objects []K
	err := Read(ctx.DB, &objects)
	log.Println(objects)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error getting all objects.")
		return
	}

	log.Println(objects)
	b, err := json.Marshal(objects)
	if err != nil {
		log.Print("Error marshalling.")
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(b))
}

func CreateHandler[T any](ctx *Ctx, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var object T
	err := decoder.Decode(&object)

	// Let the gatekeeping begin.
	if err != nil {
		log.Println("Error decoding the object from the request.")
		msg, statusCode := validateJsonError(err)
		http.Error(w, msg, statusCode)
		log.Println(msg)
		return
	}

	log.Println("Decoded object: ", object)

	// Attempting to save the object to the db.
	res, err := Create(ctx.DB, object)

	if err != nil {
		log.Print("Error saving object to db.")
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Object created!, id: %s", *res)
}

func ReplaceHandler[T any](ctx *Ctx, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var object T
	err := decoder.Decode(&object)

	// Let the gatekeeping begin.
	if err != nil {
		log.Println("Error decoding the object from the request.")
		msg, statusCode := validateJsonError(err)
		http.Error(w, msg, statusCode)
		log.Println(msg)
		return
	}

	log.Println("Decoded object: ", object)

	// Attempting to save the object to the db.
	err = ReplaceOne(ctx.DB, &object, vars["id"])
	if err != nil {
		log.Print("Error replacing object in db.")
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Object updated!")
}

func DeleteHandler[T any](ctx *Ctx, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Adding additional checks in this generic function would be difficult.
	// mongodb does not support cascade deletes.
	// If you need more validation and dependency checking, please use a seperate handler for the same.
	err := Delete[T](ctx.DB, vars["id"])
	if err != nil {
		http.Error(w, "Error deleting object.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Object deleted.")
}

func validateJsonError(err error) (string, int) {
	msg := ""
	var statusCode int
	var unmarshallError *json.UnmarshalTypeError
	var syntaxError *json.SyntaxError
	switch {
	case errors.As(err, &syntaxError):
		msg = fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
		statusCode = http.StatusBadRequest
	case errors.As(err, &unmarshallError):
		msg = fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshallError.Field, unmarshallError.Offset)
		statusCode = http.StatusBadRequest
	case errors.Is(err, io.ErrUnexpectedEOF):
		msg = "Request body contains badly-formed JSON"
		statusCode = http.StatusBadRequest
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		msg = fmt.Sprintf("Request body contains unknown field %s", fieldName)
		statusCode = http.StatusBadRequest
	case errors.Is(err, io.EOF):
		msg = "Request body must not be empty"
		statusCode = http.StatusBadRequest
	case err.Error() == "http: request body too large":
		msg = "Request body must not be larger than 1MB"
		statusCode = http.StatusRequestEntityTooLarge
	default:
		msg = http.StatusText(http.StatusInternalServerError)
		statusCode = http.StatusInternalServerError
	}
	return msg, statusCode
}
