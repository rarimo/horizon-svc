package msgs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	raw := `{"raw":{"start":1734,"end":1735},"type":"block_range"}`

	m := Message{raw: []byte(raw)}

	if !assert.Equal(t, MessageTypeBlockRange, m.Type()) {
		return
	}

	brm := m.MustBlockRangeMessage()

	if !assert.Equal(t, int64(1734), brm.Start) {
		return
	}

	if !assert.Equal(t, int64(1735), brm.End) {
		return
	}
}
