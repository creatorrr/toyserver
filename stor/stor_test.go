// Tests for stor package.

package stor_test

import (
	"math/rand"

	. "github.com/creatorrr/toyserver/stor"
	"testing"
)

const (
	key string = "alice"
)

var m = &Model{
	key,
	make(map[string]interface{}),
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

func TestModel(t *testing.T) {
	// Make sure model has correct collection name.
	coll := m.Collection()
	if coll != "Sessions" {
		t.Errorf("Wrong collection name", string(coll))
	}

	// Set up model and save it.
	dat := randStr(25)
	m.Data["str"] = dat

	// Blocking call.
	if !(<-m.Save()) {
		t.Errorf("Model not saved.")
		return
	}

	// Now reset model and get value.
	delete(m.Data, "str")
	if !(<-m.Get()) {
		t.Errorf("Model not fetched.")
	}

	if m.Data["str"] != dat {
		t.Errorf("Incorrect data:", m.Data["str"])
	}

	// Finally delete key.
	if !(<-m.Delete()) {
		t.Errorf("Model not deleted.")
	}
}
