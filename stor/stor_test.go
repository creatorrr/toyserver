// Tests for stor package.

package stor_test

import (
	"errors"
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
	if v, err = json.Marshal(d); err != nil {
		return nil, err
	}

	return
}

func (d *ModelData) SetJson(s []byte) error {
	if e := json.Unmarshal(s, &d); e != nil {
		return e
	}

	return nil
}

// ModelData implements DataModeler interface
func (d *ModelData) GetValue() interface{} {
	return d
}

func (d *ModelData) SetValue(v interface{}) error {
	if k, e := v.(*ModelData); e {
		*d = *k
		return nil
	}

	return errors.New("cannot convert value to ModelData")
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
		_ Modeler     = &Model{}
		_ DataModeler = &ModelData{}
	)

	// Make sure model has correct collection name.
	coll := m.Collection()
	if coll != "Sessions" {
		t.Errorf("Wrong collection name", string(coll))
	}

	// Set up model and save it.
	dat := randStr(25)
	e = m.Data.SetValue(&ModelData{
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
	m.Data.SetValue(&ModelData{})
	if e = <-m.Get(); e != nil {
		t.Errorf("Model not fetched.")
		return
	}

	v := m.Data.GetValue().(*ModelData)
	if (*v).AppData["str"] != dat {
		t.Errorf("Incorrect data:", (*v).AppData["str"])
		return
	}

	// Finally delete key.
	if e = <-(m.Delete()); e != nil {
		t.Errorf("Model not deleted.")
		return
	}

	// Shut down.
	Shutdown()
}

// TODO: Add simple benchmarks for operations.
