package deepcopy

import (
	"bytes"
	"encoding/gob"
)

// Map returns a deep copy of a map.
func Map(m map[string]interface{}) map[string]interface{} {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	err := enc.Encode(m)
	if err != nil {
		panic(err)
	}

	var copy map[string]interface{}
	err = dec.Decode(&copy)
	if err != nil {
		panic(err)
	}

	return copy
}

func init() {
	gob.Register(map[string]interface{}{})
}
