// Package stor provides a simple dal, creates models and abstract
// access to backend data storage provider, in this case orchestrate.io.

package stor

import (
	"log"
	"os"
	"strings"

	"encoding/json"

	orchestrate "github.com/creatorrr/orchestrate-go-client"
)

// Define interfaces.
type (
	jsoner interface {
		Json() ([]byte, error)
		SetValue([]byte) error
	}

	Modeler interface {
		jsoner
		Collection() string

		Get() <-chan error
		Save() <-chan error
		Delete() <-chan error
	}

	worker interface {
		Work()
	}
)

// Define data types.
type (
	User struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	SessionData struct {
		AppData map[string]interface{} `json:"appData"`
		Members []User                 `json:"members"`
	}

	Session struct {
		Key  string
		Data *SessionData
		Type string
	}
)

// Define work and worker.
const (
	GET = iota
	PUT
	DELETE
)

// TODO: Make Payload accept Modeler
type (
	work struct {
		Type    int
		Payload *Session
		Notif   *chan error
	}

	transaction struct {
		Queue chan *work
	}
)

// Set up client and event loop.
var (
	dal       *orchestrate.Client
	workQueue chan *work
)

func init() {
	apiKey := os.Getenv("ORCHESTRATE_API_KEY")
	if apiKey == "" {
		panic("Api key not found.")
	}

	dal = orchestrate.NewClient(apiKey)
	workQueue = make(chan *work, 1)

	// TODO: Abstract the goroutine to expose a Start() method on package.
	// Set up event loop.
	go func() {
		trs := make(map[string]*transaction)

		for w := range workQueue {
			m := w.Payload
			trKey := m.Collection() + "/" + m.Key

			// Add new transaction if it doesn't exist.
			if _, ok := trs[trKey]; !ok {
				trs[trKey] = &transaction{
					make(chan *work, 1),
				}

				go trs[trKey].Work()
			}

			// Push work to new transaction.
			trs[trKey].Queue <- w
		}

		// Clean up.
		for _, tr := range trs {
			close(tr.Queue)
		}
	}()
}

func Shutdown() {
	if workQueue != nil {
		close(workQueue)
	}

	// Add other cleanup code here.
}

// transaction implements worker interface
func (t *transaction) Work() {
	// TODO: timeout
	for w := range t.Queue {
		m := w.Payload

		switch w.Type {
		case PUT:
			val, _ := m.Json()

			*w.Notif <- dal.Put(m.Collection(), m.Key, strings.NewReader(string(val)))
			close(*w.Notif)

		case DELETE:
			*w.Notif <- dal.Delete(m.Collection(), m.Key)
			close(*w.Notif)
			// case GET: // Not implemented.
		}
	}
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
		val, e := dal.Get(m.Collection(), m.Key)

		// Set data value.
		jsonE := m.SetValue([]byte(val.String()))

		if e != nil {
			q <- e

		} else {
			// Send notification on channel.
			q <- jsonE
		}
	}()

	return q
}

func (m *Session) Save() <-chan error {
	defer catch()

	q := make(chan error, 1)

	// Feed to work queue.
	workQueue <- &work{
		PUT,
		m,
		&q,
	}

	return q
}

func (m *Session) Delete() <-chan error {
	defer catch()

	q := make(chan error, 1)

	// Feed to work queue.
	workQueue <- &work{
		DELETE,
		m,
		&q,
	}

	return q
}
