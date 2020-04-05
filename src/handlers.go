package main

import (
	"encoding/json"
	"net/http"

	"github.com/samziz/godis/src/database"
)

// This is another file but it's also in the `main` package, so
// you can basically treat it as if it's the same file as `main.go`.
// You don't need to explicitly import anything from here into `main.go`,
// it's just available automatically.

// This is our get handler. It just gets a Value that the user has requested.
//
// You can also define functions like this, as variables. This is useful
// I guess if you want to do clever stuff involving reassigning their Value.
// We don't in this case: I just wanted to define it at this point for the
// sake of narrative order, and I can't do that using the normal syntax.
func handleGet(reqBody Request, w http.ResponseWriter, db database.Database) {
	var rsp Response // create a Response variable - currently zero valued
	 				 // this syntax is the exact same as doing `rsp := Response{}`

	value, err := db.Get(reqBody.Key)
	if err != nil {
		// Write a 500 (internal server error) response with the raw error.
		// We wouldn't do this in production, again, but this is an example!

		// Get `rsp` (defined above) and assign to it a Response with an Error
		// field and a Status of 500 (internal server Error).
		rsp = Response{
			Error:  err,
			Status: 500,
		}
	} else {
		// There was no error. Assign a successful response.
		rsp = Response{
			Status: 200,
			Value:  value,
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
	// we call it a 200 even if we didn't find the key, just because
	// technically the database query was successful, it just didn't
	// find anything. Feel free to disagree with me here.
	w.Write(jsonBytes)
	return
}

// This is our set handler. It just sets a value and then responds 'ok'.
func handleSet(reqBody Request, w http.ResponseWriter, db database.Database) {
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
		Status: 200,
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