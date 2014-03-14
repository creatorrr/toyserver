// Package stor provides a simple DAL, creates models and abstract
// access to backend data storage provider, in this case orchestrate.io.

package stor

import (
	"log"
	"os"
	"strings"

	"encoding/json"

	orchestrate "github.com/orchestrate-io/orchestrate-go-client"
)

// Set up orchestrate client
var DAL *orchestrate.Client

func init() {
	apiKey := os.Getenv("ORCHESTRATE_API_KEY")
	if apiKey == "" {
		panic("Api key not found.")
	}

	DAL = orchestrate.NewClient(apiKey)
}

// Define data types.
type Data map[string]interface{}

type Model struct {
	Key  string
	Data Data
	Type string
}

type Modeler interface {
	Collection() string

	Get() chan bool
	Save() chan bool
	Delete() chan bool
}

type Jsoner interface {
	Json() []byte
	SetValue([]byte) bool
}

// Model implements Jsoner interface
func (m *Model) Json() (v []byte, err error) {
	if v, err = json.Marshal(m.Data); err != nil {
		return nil, err
	}

	return
}

func (m *Model) SetValue(s []byte) error {
	if e := json.Unmarshal(s, &m.Data); e != nil {
		return e
	}

	return nil
}

// Model implements Modeler interface

// catch must be run deferred so it can recover from runtime panics
func catch() {
	if r := recover(); r != nil {
		log.Println("Panic occurred:", r)
	}
}

func (m *Model) Collection() (a string) {
	a = m.Type

	// Capitalize name.
	a = strings.ToUpper(string(a[0])) + a[1:]

	// Pluralize (naively).
	a += "s"
	return
}

func (m *Model) Get() (q chan bool) {
	defer catch()

	q = make(chan bool, 1)

	go func() {
		defer close(q)

		// Get object from dal.
		val, e := DAL.Get(m.Collection(), m.Key)
		jsonE := m.SetValue([]byte(val.String()))

		// Send notification on channel.
		q <- (e == nil && jsonE == nil)
	}()

	return
}

func (m *Model) Save() (q chan bool) {
	defer catch()

	q = make(chan bool, 1)

	go func() {
		defer close(q)

		// Get json value of object.
		val, jsonE := m.Json()
		e := DAL.Put(m.Collection(), m.Key, strings.NewReader(string(val)))

		// Send notification on channel.
		q <- (e == nil && jsonE == nil)
	}()

	return
}

func (m *Model) Delete() (q chan bool) {
	defer catch()

	q = make(chan bool, 1)

	go func() {
		defer close(q)

		// Delete key.
		e := DAL.Put(m.Collection(), m.Key, strings.NewReader("{}"))

		// Send notification on channel.
		q <- (e == nil)
	}()

	return
}
