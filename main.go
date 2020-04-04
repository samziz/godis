package main

import (
	// These imports are all from the Go standard library.
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	// This imports `database` from the `./database` directory.
	"github.com/samziz/godis/database"
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
	// IS typed though - you can't later assign a Value of another type to it.
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
		// This is our request type. Defining it here is the exact same as defining
		// it outside of this whole function - I just wanted it to be closer to the
		// action for the sake of intelligibility.
		type Request struct {
			// Let's accept an `Op` field which is just what the user wants to
			// do: get or set (i.e. read or write).
			Op string

			Key string

			// This will be empty if the Op is "GET".
			Value string
		}

		// Let's parse the request.

		// First, create an instance of Request. This is empty. An empty struct has
		// a 'zero Value' for each of its fields: so 0 if it's a number type, and
		// "" if it's a string. Go doesn't have null or anything like that.
		//
		// However, lots of people abuse pointers to get kinda-null values. So you
		// can define a field as a pointer, call `json.Unmarshal` on it, and it will
		// be left as a nil pointer if there was no Value: so you can tell between
		// 0 and no Value for an int field, for example.
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
			// function, one of which is an error. You've gotta check that error Value and
			// do something about it. In this case we'll return an HTTP error.

			// Also, this is how you set a status code on the response. If you don't set one,
			// it's 200 (OK) by default.
			w.WriteHeader(400)

			// In reality you might want to write your own slightly friendlier error
			// rather than leak ugly internal details to the end user.
			w.Write([]byte(err.Error()))

			return
		}

		// This is our response type. Same as above, could have defined it elsewhere
		// but wanted to do it here so you could read the code like a narrative.
		type Response struct {
			// This will be missing if there is no error.
			error error

			status uint

			// This will be missing if there IS an error.
			value string
		}

		// This is our get handler. It just gets a Value that the user has requested.
		//
		// You can also define functions like this, as variables. This is useful
		// I guess if you want to do clever stuff involving reassigning their Value.
		// We don't in this case: I just wanted to define it at this point for the
		// sake of narrative order, and I can't do that using the normal syntax.
		var handleGet = func(reqBody Request, w http.ResponseWriter, db database.Database) {
			var rsp Response // create a Response variable - currently zero valued

			value, err := db.Get(reqBody.Key)
			if err != nil {
				// Write a 500 (internal server error) response with the raw error.
				// We wouldn't do this in production, again, but this is an example!

				// Get `rsp` (defined above) and assign to it a Response with an error
				// field and a status of 500 (internal server error).
				rsp = Response{
					error: err,
					status: 500,
				}
			} else {
				// There was no error. Assign a successful response.
				rsp = Response{
					status: 200,
					value:  value,
				}
			}

			jsonBytes, err := json.Marshal(rsp)
			if err != nil {
				// Freak out and just write a plain 500.
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}

			// Write a response and exit. There's a bit of weirdness here -
			// we call it a 200 even if we didn't find the Key, just because
			// technically the database query was successful, it just didn't
			// find anything. Feel free to disagree with me here.
			w.Write(jsonBytes)
			return
		}

		// This is our set handler. It just sets a Value and then responds 'ok'.
		var handleSet = func(reqBody Request, w http.ResponseWriter, db database.Database) {
			err := db.Set(reqBody.Key, reqBody.Value)
			if err != nil {
				// Freak out and just write a plain 500. Unlike the get handler, this
				// IS a regular old HTTP error, because the program failed to do its job.
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}

			// There was no error. Assign a successful response.
			rsp := Response{
				status: 200,
			}

			jsonBytes, err := json.Marshal(rsp)
			if err != nil {
				// Freak out and just write a plain 500.
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}

			w.Write(jsonBytes)
			return
		}

		// Switch on the Value of `Op`, handling gets and sets differently.
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
			w.Write([]byte("Unrecognised Op: " + reqBody.Op))
		}
	})

	// Now we've defined our handler, let's start listening. Passing nil
	// as the second param means it uses the DefaultServeMux variable,
	// which has been configured with the handler we wrote just above.
	// Also the fmt.Sprintf call returns the format string (the first argument)
	// with "%s" replaced with the second argument, `port`.
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%s", port), nil)
}