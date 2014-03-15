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

// Define interfaces.
type Modeler interface {
	Collection() string

	Get() <-chan error
	Save() <-chan error
	Delete() <-chan error
}

type Jsoner interface {
	Json() ([]byte, error)
	SetValue([]byte) error
}

// Define data types.
type User struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type SessionData struct {
	AppData map[string]interface{} `json:"appData"`
	Members []User                 `json:"members"`
}

type Session struct {
	Key  string
	Data *SessionData
	Type string
}

// Model implements Jsoner interface
func (m *Session) Json() (v []byte, err error) {
	if v, err = json.Marshal(m.Data); err != nil {
		return nil, err
	}

	return
}

func (m *Session) SetValue(s []byte) error {
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

func (m *Session) Collection() (a string) {
	a = m.Type

	// Capitalize name.
	a = strings.ToUpper(string(a[0])) + a[1:]

	// Pluralize (naively).
	a += "s"
	return
}

func (m *Session) Get() <-chan error {
	defer catch()

	q := make(chan error, 1)

	go func() {
		defer close(q)

		// Get object from dal.
		val, e := DAL.Get(m.Collection(), m.Key)

		if e != nil {
			q <- e
			return
		}

		// Set data value.
		jsonE := m.SetValue([]byte(val.String()))

		// Send notification on channel.
		q <- jsonE
	}()

	return q
}

func (m *Session) Save() <-chan error {
	defer catch()

	q := make(chan error, 1)

	go func() {
		defer close(q)

		// Get json value of object.
		val, jsonE := m.Json()

		if jsonE != nil {
			q <- jsonE
			return
		}

		// Save value.
		e := DAL.Put(m.Collection(), m.Key, strings.NewReader(string(val)))

		// Send notification on channel.
		q <- e
	}()

	return q
}

func (m *Session) Delete() <-chan error {
	defer catch()

	q := make(chan error, 1)

	go func() {
		defer close(q)

		// Delete key.
		e := DAL.Put(m.Collection(), m.Key, strings.NewReader("{}"))

		// Send notification on channel.
		q <- e
	}()

	return q
}
