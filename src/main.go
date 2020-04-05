package main

import (
	// These imports are all from the Go standard library.
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	// This imports `database` from the `./database` directory.
	"github.com/samziz/godis/src/database"
)

// The `main` function is the thing that gets run in a Go program.
// This is your entrypoint - everything starts here.
func main() {
	// Let's do our setup here. At a bare minimum, the user should be able
	// to specify what port they want the database to listen on. The
	// `foo := doSomething()` syntax is just the same as writing
	// `var foo = doSomething()`. With the var syntax (not with the :=
	// syntax) you CAN specify a type (by doing `var foo string = doSomething()`)
	// but you don't really need to - the compiler will infer it. The variable
	// IS typed though - you can't later assign a value of another type to it.
	port := os.Getenv("PORT")

	// The little {} bit is the syntax for creating an instance of a struct.
	// This gives us a new MapDatabase instance that we can call Get() and
	// Set() on.
	db := database.NewMapDatabase()

	listen(port, &db)
}

// This `listen` method takes a `database.Database` and starts up a web server,
// spinning forever and serving requests.
func listen(port string, db database.Database) {
	// This registers a Handler with the one global instance of DefaultServeMux
	// in the `http` package. Now any subsequent call to `http.ListenAndServe`
	// will use this function as a handler. Yes, I know, no sane language would
	// do this because good programmers know not to abuse global mutable state.
	// It should return a function so we can instantiate other different HTTP
	// servers with different handlers. But it doesn't, because the Go devs
	// are a bit stubborn and weird.
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Let's parse the request.

		// First, create an instance of Request. This is empty. An empty struct has
		// a 'zero value' for each of its fields: so 0 if it's a number type, and
		// "" if it's a string. Go doesn't have null or anything like that.
		//
		// However, lots of people abuse pointers to get kinda-null values. So you
		// can define a field as a pointer, call `json.Unmarshal` on it, and it will
		// be left as a nil pointer if there was no value: so you can tell between
		// 0 and no value for an int field, for example.
		reqBody := Request{}

		// This basically does the same as `JSON.parse()` in JavaScript, or `json.dump`
		// in Python. It takes a pointer to `reqBody` and parses `req`, assigning all the
		// values to the corresponding fields in `Request`. It works out which are which
		// based on the field names, but you can use struct tags if you ever want to
		// assign a JSON field to a struct field of a different name:
		// see here for more -> https://gobyexample.com/json.
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqBody)
		if err != nil {
			// Just a note on error handling since this is the first time we've encountered
			// it. This is how you do errors in Go - by returning multiple values from a
			// function, one of which is an error. You've gotta check that error value and
			// do something about it. In this case we'll return an HTTP error.

			// Also, this is how you set a status code on the response. If you don't set one,
			// it's 200 (OK) by default.
			w.WriteHeader(400)

			// In reality you might want to write your own slightly friendlier error
			// rather than leak ugly internal details to the end user.
			w.Write([]byte(err.Error()))

			return
		}

		// Switch on the value of `Op`, handling gets and sets differently.
		// Note that unlike other languages you don't need to call 'break' -
		// Go breaks by default and won't fall through to the next case unless
		// you explicitly call 'fallthrough'.
		switch reqBody.Op {
		case "GET":
			handleGet(reqBody, w, db)
		case "SET":
			handleSet(reqBody, w, db)
		default:
			w.WriteHeader(500)
			w.Write([]byte("Unrecognised op: " + reqBody.Op))
		}
	})

	// Now we've defined our handler, let's start listening. Passing nil
	// as the second param means it uses the DefaultServeMux variable,
	// which has been configured with the handler we wrote just above.
	// Also the fmt.Sprintf call returns the format string (the first argument)
	// with "%s" replaced with the second argument, `port`.
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%s", port), nil)
}