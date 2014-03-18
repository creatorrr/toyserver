// Package stor provides a simple dal, creates models and abstract
// access to backend data storage provider, in this case orchestrate.io.

package stor

import (
	"log"
	"os"
	"strings"
	"time"

	orchestrate "github.com/creatorrr/orchestrate-go-client"
)

// Define interfaces.
type (
	Jsoner interface {
		Json() ([]byte, error)
		SetJson([]byte) error
	}

	Modeler interface {
		Collection() string

		Get() <-chan error
		Save() <-chan error
		Delete() <-chan error
	}
)

// Define data types.
type Model struct {
	Key  string
	Data Jsoner
	Type string
}

// Define work and worker.
const (
	GET = iota
	PUT
	DELETE
)

type (
	work struct {
		Type    int
		Payload *Model
		Notif   *chan error
	}

	transaction struct {
		Queue   chan *work
		working bool
	}
)

// Set up client and event loop.
var (
	dal       *orchestrate.Client
	workQueue chan *work

	workerLifeSpan = time.Hour * 2
)

func init() {
	apiKey := os.Getenv("ORCHESTRATE_API_KEY")
	if apiKey == "" {
		panic("Api key not found.")
	}

	dal = orchestrate.NewClient(apiKey)
	workQueue = make(chan *work, 1)

	// Set up event loop.
	go Start()
}

func Start() {
	var tr *transaction
	trs := make(map[string]*transaction)

	// Distribute work
	for w := range workQueue {
		m := w.Payload
		trKey := m.Collection() + "/" + m.Key

		// Add new transaction if it doesn't exist.
		if _, ok := trs[trKey]; !ok {
			trs[trKey] = &transaction{
				make(chan *work, 100),
				false,
			}
		}

		tr = trs[trKey]

		// Start transaction goroutine.
		if !tr.working {
			go tr.Work()
		}

		// Push work to new transaction.
		tr.Queue <- w
	}

	// Clean up.
	for _, tr := range trs {
		close(tr.Queue)
	}
}

func Shutdown() {
	defer close(workQueue)

	// Add other cleanup code here.
}

// transaction implements worker interface
func (t *transaction) Work() {
	timeout := time.After(workerLifeSpan)

	for w := range t.Queue {
		m := w.Payload

		switch w.Type {
		case GET:
			// Get object from dal.
			val, e := dal.Get(m.Collection(), m.Key)

			if e == nil {
				m.Data.SetJson([]byte(val.String()))
			}

			*w.Notif <- e

		case PUT:
			val, _ := m.Data.Json()

			*w.Notif <- dal.Put(m.Collection(), m.Key, strings.NewReader(string(val)))

		case DELETE:
			*w.Notif <- dal.Delete(m.Collection(), m.Key)
		}

		// Close notification channel.
		close(*w.Notif)

		// Timeout goroutine to auto destroy after lifespan.
		select {
		case <-timeout:
			t.working = false
			return

		default:
			continue
		}
	}
}

// Model implements Modeler interface

// catch must be run deferred so it can recover from runtime panics
func catch() {
	if r := recover(); r != nil {
		defer os.Exit(1)

		log.Println("Panic occurred:", r)
		Shutdown()
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

func (m *Model) Get() <-chan error {
	defer catch()

	q := make(chan error, 1)

	// Feed to work queue.
	workQueue <- &work{
		GET,
		m,
		&q,
	}

	return q
}

func (m *Model) Save() <-chan error {
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

func (m *Model) Delete() <-chan error {
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
