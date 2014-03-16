// Tests for stor package.

package stor_test

import (
	"math/rand"
	"time"

	. "github.com/creatorrr/toyserver/stor"
	"testing"
)

const (
	key string = "alice"
)

var m = &Session{
	key,
	&SessionData{
		make(map[string]interface{}),
		make([]User, 5),
	},
	"session",
}

// Utils:
// randInt generates a random integer between min and max.
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// randStr generates a random string of given length
func randStr(l int) string {
	bytes := make([]byte, l)

	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}

	return string(bytes)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// TODO: Test non blocking operations.
func TestSession(t *testing.T) {
	var (
		e error

		// Make sure Session implements Jsoner and Modeler.
		_ Modeler = &Session{}
	)

	// Make sure model has correct collection name.
	coll := m.Collection()
	if coll != "Sessions" {
		t.Errorf("Wrong collection name", string(coll))
	}

	// Set up model and save it.
	dat := randStr(25)
	m.Data.AppData["str"] = dat

	// Blocking call.
	if e = <-m.Save(); e != nil {
		t.Errorf("Session not saved.")
		return
	}

	// Now reset model and get value.
	delete(m.Data.AppData, "str")
	if e = <-m.Get(); e != nil {
		t.Errorf("Session not fetched.")
		return
	}

	if m.Data.AppData["str"] != dat {
		t.Errorf("Incorrect data:", m.Data.AppData["str"])
		return
	}

	// Finally delete key.
	if e = <-(m.Delete()); e != nil {
		t.Errorf("Session not deleted.")
		return
	}

	// Shut down.
	Shutdown()
}

// TODO: Add simple benchmarks for operations.
