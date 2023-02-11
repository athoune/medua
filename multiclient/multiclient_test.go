package multiclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMulticlient(t *testing.T) {
	m := New(5 * 1024)
	a1 := m.LazyClient("a")
	a1.Timeout = 13 * time.Second
	a2 := m.LazyClient("a")
	assert.Equal(t, a2.Timeout, 13*time.Second)
	b := m.LazyClient("b")
	b.Timeout = 17 * time.Second
	assert.NotEqual(t, a1, b)
}
