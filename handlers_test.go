package grf_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	grf "github.com/Jyothis-P/go-rest-framework"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// Create a model
// Make sure to give json and bson structs as necessary.
type Todo struct {
	Id        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Completed bool               `json:"completed" bson:"completed"`
}

var testTodo Todo = Todo{
	Title:     "Finish testing this.",
	Completed: false,
}

func TestCreateHandler(t *testing.T) {

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Create todo test", func(mt *mtest.T) {

		jsonTodo, err := json.Marshal(testTodo)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("POST", "http://localhost:8001/todo/", bytes.NewReader(jsonTodo))
		if err != nil {
			t.Fatal(err)
		}
		res := httptest.NewRecorder()

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		appContext := grf.Ctx{
			DB: mt.DB,
		}
		grf.CreateHandler[Todo](&appContext, res, req)

		// Regex pattern to match the string with a changeable ObjectID
		pattern := `Object created!, id: \{ObjectID\("[0-9a-fA-F]{24}"\)}`

		// Compile the regex pattern
		r, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Println("There was an error compiling the regex pattern:", err)
			return
		}

		act := res.Body.String()
		exp := `Object created!, id: \{ObjectID\("[0-9a-fA-F]{24}"\)}`
		fmt.Println("Res:", act)
		// Check if the pattern matches the string
		match := r.MatchString(res.Body.String())
		if !match {
			t.Fatalf("Response is %s. Expected: %s", act, exp)
		}
	})
}
