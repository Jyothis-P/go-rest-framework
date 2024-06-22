package grf

import (
	"context"
	"log"
	"os"
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var db *mongo.Database

func TestGetPluralTableDriven(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"pointer plural test", "*models.Todo", "todos"},
		{"pointer slice plural test", "*[]models.Todo", "todos"},
		{"base plural test", "models.Todo", "todos"},
		{"es plural test", "*models.Box", "boxes"},
		{"package plural test", "main.Box", "boxes"},
	}
	// The execution loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := getPlural(tt.input)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestReadTodo(t *testing.T) {
	var result []Todo
	err := Read(db, &result)
	if err != nil {
		t.Fatalf("Failed to read from the database: %v", err)
		return
	}

	if len(todos) != len(result) {
		t.Fatalf("Length mismatch. Expected lenght: %d. Result length: %d", len(todos), len(result))
		return
	}

	for _, todo := range todos {
		idx := slices.IndexFunc(result, func(t Todo) bool { return t.Title == todo.(Todo).Title })
		if idx == -1 {
			t.Fatalf("Expected Todo not found in db: %s", todo.(Todo).Title)
			return
		}
		if result[idx].Completed != todo.(Todo).Completed {
			t.Fatalf("Todo completed mismatch. Todo: %s. Expected: %s. Result: %s", todo.(Todo).Title, todo.(Todo).isCompleted(), result[idx].isCompleted())
			return
		}
	}
}

func TestReadOneTodo(t *testing.T) {
	for _, tt := range todos {
		expected := tt.(Todo)
		t.Run("Todo: "+expected.Title, func(t *testing.T) {
			var result Todo
			err := ReadOne(db, &result, expected.Id.Hex())
			if err != nil {
				t.Fatalf("Failed to read from the database: %v", err)
				return
			}

			if expected.Id != result.Id {
				t.Fatalf("Id mismatch. Expected: %v, Result: %v", expected.Id, result.Id)
				return
			}

			if expected.Title != result.Title {
				t.Fatalf("Title mismatch. Expected: %s, Result: %s", expected.Title, result.Title)
				return
			}

			if expected.Completed != result.Completed {
				t.Fatalf("Completed mismatch. Expected: %s, Result: %s", expected.isCompleted(), result.isCompleted())
				return
			}
		})
	}
}

func TestMain(m *testing.M) {
	// Setup database.
	var cancel func()
	var err error
	log.Print("Connecting to db... ")
	db, cancel, err = SetupTestDatabase()
	defer cancel()
	if err != nil {
		log.Fatalln("Error setting up the database.", err)
		return
	}
	log.Println("Done.")
	log.Print("Filling up db... ")
	err = loremIpsumDb()
	if err != nil {
		log.Fatalln("Error filling up the database.", err)
		return
	}
	log.Println("Done.")

	// Calling the tests.
	exitCode := m.Run()

	log.Print("Cleaning up db... ")
	err = cleanUpDB()
	if err != nil {
		log.Fatalln("Error cleaning up the database.", err)
	} else {
		log.Println("Done.")
	}
	os.Exit(exitCode)
}

type Todo struct {
	Id        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Completed bool               `json:"completed" bson:"completed"`
}

func (t Todo) isCompleted() string {
	if t.Completed {
		return "Complete"
	} else {
		return "Incomplete"
	}
}

type Box struct {
	Id     primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Length int                `json:"length" bson:"length"`
	Width  int                `json:"width" bson:"width"`
	Height int                `json:"height" bson:"height"`
}

// ==== Helper functions to fill and scrub the test db for testing. ====

var todoTitles []string = []string{
	"Feed the dogs",
	"Take them for a walk",
	"Clean the house",
	"Dream big",
}

var todos []interface{}

func loremIpsumDb() error {
	todos = make([]interface{}, 4)
	for i, todo := range todoTitles {
		todos[i] = Todo{
			Id:        primitive.NewObjectID(),
			Title:     todo,
			Completed: (i % 2) == 0,
		}
	}

	collection, ctx, cancel := getCollectionAndContext[Todo](db, todos[0].(Todo))
	defer cancel()

	_, err := collection.InsertMany(ctx, todos)
	if err != nil {
		log.Fatalln("Error adding object to database.", err)
		return err
	}
	return nil
}

func cleanUpDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	collection := getCollection[Todo](db)
	err := cleanUpCollection(*collection, ctx)
	return err
}

func cleanUpCollection(collection mongo.Collection, ctx context.Context) error {
	filter := bson.D{}
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Fatalln("Error cleaning up the collection.", collection.Name(), err)
		return err
	}
	return nil
}
