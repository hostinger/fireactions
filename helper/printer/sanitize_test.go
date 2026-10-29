package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	assert.Equal(t, "plain text ąčę", Sanitize("plain text ąčę"))
	assert.Equal(t, `\x1b[2Jcleared`, Sanitize("\x1b[2Jcleared"))
	assert.Equal(t, `a\x0ab\x09c`, Sanitize("a\nb\tc"))
	assert.Equal(t, "a\nb\tc", Sanitize("a\nb\tc", '\n', '\t'))
	assert.Equal(t, `\x7f\x9b\x00`, Sanitize("\x7f\u009b\x00"))
}

type fakePrintable struct{}

func (fakePrintable) Cols() []string { return []string{"Name", "State"} }

func (fakePrintable) ColsMap() map[string]string {
	return map[string]string{"Name": "Name", "State": "State"}
}

func (fakePrintable) KV() []map[string]interface{} {
	return []map[string]interface{}{{"Name": "runner-1", "State": "\x1b[2J\x1b[Hidle"}}
}

func TestPrintTextEscapesControlCharacters(t *testing.T) {
	var out bytes.Buffer
	PrintText(fakePrintable{}, &out, nil)

	assert.Contains(t, out.String(), `\x1b[2J\x1b[Hidle`)
	assert.NotContains(t, out.String(), "\x1b")
}
