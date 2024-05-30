# go-rest-framework

An opinionated Go web framework with MongoDB support for rapid REST API development. Simply define your models and get started with your basic CRUD routes.

## Features

1. Provides a new Handler adapter which can passes an Application Context to your handler functions. This way,  you can access. your DB from any package without a lot of global variables.
    
    Based on this article: ***[Custom Handlers and Avoiding Globals in Go Web Applications](https://blog.questionable.services/article/custom-handlers-avoiding-globals/)***
    
2. Provides ready to use generic services for CRUD from your mongodb. 
3. Provides ready to use generic handlers for CRUD from your mongodb. 

## Getting started

1. Environment setup.
    
    GRF expects that you have mongodb set up with Replicasets in your environment. This is to avail the transaction capabilities of the database and ensure ACID transactions. If you're looking for a quick and easy way to set it up, I'd recommend checking out [run-rs](https://www.npmjs.com/package/run-rs).
    
    Once the db is setup, please add the following variables to your .env file at the project root.
    
    ```bash
    DATABASE_URI=<Your mongodb connection string>
    DATABASE_NAME=<Your database name>
    ```
    
2. Create an App Context with a DB connection.
    
    We need a *mongo.Database object that we can pass in the Application context to the handlers. You can use the SetupDatabase() function to ease this up.
    
    ```go
    import grf "github.com/Jyothis-P/go-rest-framework"
    
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
    ```
    
3. Create a router. [Currently modelled for gorilla/mux will be generalising to add support for anything that supports http/Handler]
    
    ```go
    // Create a router.
    r := mux.NewRouter().StrictSlash(true)
    ```
    
4. Create your model.
    
    Make sure to add json and bson struct tags to help configure marshalling. Showing a model to store TODO objects. bson struct tags are used for attribute names in the db. json struct tags are used to determine the json parameters during http requests. 
    
    ```go
    // Create a model
    // Make sure to give json and bson structs as necessary.
    type Todo struct {
    	Title     string `json:"title" bson:"title"`
    	Completed bool   `json:"completed" bson:"completed"`
    }
    ```
    
5. Register the CRUD routes.
    
    RegisterCRUDRoutes is a generic that takes the model as the type parameter. It creates a subrouter with the given path. This router can be optionally returned if further routes are need on it.
    
    ```go
    // Register routes for the model.
    grf.RegisterCRUDRoutes[Todo]("/todo", r, &appContext)
    ```
    
    That’s it! This function will take care of the services and handlers required for all the basic CRUD REST endpoints for your model. 
    
    The data will be saved in a collection with the plural form of your model’s name. The collection name will be `todos` for this example.
    
6. Setup and start the webserver.
    
    ```go
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
    ```
    

## TODO

- [ ]  DB agnostic
- [ ]  Remove dependency from .env file.
- [ ]  Tests