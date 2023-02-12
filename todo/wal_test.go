package todo

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWal(t *testing.T) {
	dir, err := os.MkdirTemp("", "dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up
	o, err := os.OpenFile(filepath.Join(dir, "demo.wal"), os.O_CREATE+os.O_WRONLY, 0600)
	assert.NoError(t, err)
	t1, err := NewWithWal(o, 32)
	assert.NoError(t, err)
	n, err := t1.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	n, err = t1.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
	err = t1.Reset(0)
	assert.NoError(t, err)
	err = o.Close()
	assert.NoError(t, err)
	o2, err := os.OpenFile(filepath.Join(dir, "demo.wal"), os.O_RDWR, 0600)
	assert.NoError(t, err)
	t2, err := ReadWal(o2)
	assert.NoError(t, err)
	assert.Equal(t, int64(32), t2.size)
	n, err = t2.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}
