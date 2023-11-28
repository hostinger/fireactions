package stringid

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
)

const (
	// StringIDLength is the length of a string ID.
	StringIDLength = 12
)

// New returns a new string ID.
func New() string {
	b := make([]byte, StringIDLength)
	for {
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}

		id := hex.EncodeToString(b)
		_, err := strconv.ParseInt(id, 10, 64)
		if err == nil {
			continue
		}

		return id
	}
}
