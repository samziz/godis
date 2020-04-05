package database

import (
	"errors"
	"sync"
)

// Let's define a Database interface. An interface basically is something
// you can define a function to accept (see `writeToDatabase` in `main.go`)
// and then you can pass any *type* that implements that interface to the
// function.
//
// The `error` interface is pretty common as the way to communicate errors
// in Go - it's just an interface that can be satisfied by any type that
// has an Error() method (see https://golang.org/pkg/builtin/#error).
// There's nothing special about an `error` - you could write a similar
// interface yourself and have all your error types implement it, and all
// your functions return it.
//
// Our Database should be a hashtable database like Redis - basically a
// hashtable (like a Python dictionary) that can be written to and read
// from by remote clients over HTTP. So that's what our interface defines
// - a function to read a string value by a string key, and a function
// to set a string value by a string key.
type Database interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

// This is our implementation of the Database interface. It would be
// possible to also write another implementation that uses something
// different (say, our own hashtable implemented in plain Go, not
// using the map type) and we could use that instead in any function
// that accepts a type that satisfies the `Database` interface.
//
// Note that MapDatabase satisfies the Database interface implicitly.
// i.e. unlike some languages, it doesn't have to SAY "I satisfy the
// Database interface". The compiler just knows that it's ok to pass
// a MapDatabase to any function that accepts a Database because it
// can see that a MapDatabase has all the methods it needs in order
// to count as a Database.
//
// It's pretty common to use interfaces to make sure that our code is
// decoupled: https://softwareengineering.stackexchange.com/questions/244476/what-is-decoupling-and-what-development-areas-can-it-apply-to
type MapDatabase struct {
	// This is a mutex to lock the hashtable so we can't access it
	// concurrently. With a mutex, if it's currently locked, then a
	// call to lock will spin until it's unlocked. This guarantees no
	// two threads can take a lock (and execute the code within the
	// critical section where the lock is taken) at the same time.
	//
	// This leaves an optimisation opportunity, since a map CANNOT be
	// read while written or written while written, but as long as it's
	// NOT being written, it can accept concurrent reads. Given the use
	// pattern will probably be lots of concurrent reads and fewer writes,
	// you could look at using a sync.RWMutex.
	hashtableMutex sync.Mutex

	// This is the hashtable (you know, like a Python dictionary - in Go
	// they're called maps)
	//
	// This field is unexported (in other languages you'd call this private).
	// The way you export a field is to give it a name that starts with a
	// capital letter. This is extremely stupid but it's how the language is
	// written.
	hashtable      map[string]string
}

// This is a method on a struct receiver. You can call it with an instance of
// `MapDatabase`, so like this:
//
// 		var db := MapDatabase{}
//		db.Set("foo", "bar")
//		db.Get("foo") // returns "bar"
//
// Every struct method can take either a pointer or a non-pointer receiver.
// If you take a pointer receiver, anything you do to `db` inside the method
// will affect the struct that the user calls the method on (like if you
// change the value of a field, their struct will now have a different value
// for that field). You can use `(db *MapDatabase)` instead of using
// `(db MapDatabase)` to use a pointer receiver instead, but since we
// don't need one here, I've used a non-pointer receiver just to make it clear
// to the reader that we're not mutating the struct in any way.
func (db MapDatabase) Get(key string) (string, error) {
	v, ok := db.hashtable[key]
	if !ok {
		return "", errors.New("this value doesn't exist")
	}

	return v, nil
}

// This function DOES use a pointer receiver since we want to modify the
// struct that the user calls it on. It would be no use using a value (non-ptr)
// receiver, since we wouldn't be setting the value in the same actual copy
// of the struct that the caller has. So they would call Set("foo", "bar")
// and then Get("foo") and find out that that key hasn't been set.
//
// Also note that we don't ever return an error in this case. The only reason
// this function is defined as returning an error is just to satisfy the
// `Database` interface above, since it requires that the Set() method
// must return an error variable. Ours is just always gonna be nil in this case.
func (db *MapDatabase) Set(key, value string) error {
	db.hashtableMutex.Lock()

	// This executes db.hashtableMutex.Unlock() whenever the function exits,
	// and is common for more complicated functions because it's hard to
	// identify all the possible paths where it might exit.
	defer db.hashtableMutex.Unlock()

	db.hashtable[key] = value

	return nil
}

// This is helpful because just doing `db := MapDatabase{}` will use the zero
// value for `hashtable`, which is a nil map. Assigning to a nil map causes a
// panic - you need to initialise an actual map for this field. But you can't
// do this from outside this package because that field is unexported, so we
// need this function to do it for us.
func NewMapDatabase() MapDatabase {
	return MapDatabase{
		hashtable: map[string]string{},
	}
}
