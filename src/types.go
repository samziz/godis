package main

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

// This is our response type. Same as above, could have defined it elsewhere
// but wanted to do it here so you could read the code like a narrative.
type Response struct {
	// This will be missing if there is no error.
	Error error `json:",omitempty"`

	Status uint

	// This will be missing if there IS an error,
	// or if this is a SET operation.
	Value string `json:",omitempty"`
}