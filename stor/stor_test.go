// Tests for stor package.

package stor_test

import (
	"math/rand"
	"time"

	"encoding/json"

	. "github.com/creatorrr/toyserver/stor"
	"testing"
)

const (
	key string = "alice"
)

type (
	User struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	ModelData struct {
		AppData map[string]interface{} `json:"appData"`
		Members []User                 `json:"members"`
	}
)

// ModelData implements Jsoner interface
func (d *ModelData) Json() (v []byte, err error) {
	v, err = json.Marshal(d)
	return
}

func (d *ModelData) SetJson(s []byte) (e error) {
	e = json.Unmarshal(s, &d)
	return
}

var m = &Model{
	key,
	&ModelData{
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
func TestModel(t *testing.T) {
	var (
		e error

		// Make sure Model implements Jsoner and Modeler.
		_ Modeler = &Model{}
		_ Jsoner  = &ModelData{}
	)

	// Make sure model has correct collection name.
	coll := m.Collection()
	if coll != "Sessions" {
		t.Errorf("Wrong collection name", coll)
	}

	// Set up model and save it.
	dat := randStr(25)
	m.Data = Jsoner(&ModelData{
		map[string]interface{}{"str": dat},
		make([]User, 5),
	})

	if e != nil {
		t.Errorf("ModelData not set:", e)
		return
	}

	if _, e = m.Data.Json(); e != nil {
		t.Errorf("json parsing error:", e)
		return
	}

	// Blocking call.
	if e = <-m.Save(); e != nil {
		t.Errorf("Model not saved.")
		return
	}

	// Now reset model and get value.
	m.Data = Jsoner(&ModelData{})
	if e = <-m.Get(); e != nil {
		t.Errorf("Model not fetched.")
		return
	}

	v := m.Data.(*ModelData)
	if v.AppData["str"] != dat {
		t.Errorf("Incorrect data:", v.AppData["str"])
		return
	}

	// Finally delete key.
	if e = <-m.Delete(); e != nil {
		t.Errorf("Model not deleted.")
		return
	}

	// Shut down.
	Shutdown()
}

// TODO: Add simple benchmarks for operations.
