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
	o, err := os.OpenFile(filepath.Join(dir, "demo.wal"), os.O_CREATE+os.O_RDWR, 0600)
	assert.NoError(t, err)
	t1, err := ReadFromWal(o, 32)
	assert.NoError(t, err)
	n := t1.Next()
	assert.Equal(t, int64(0), n)
	err = t1.Done(n)
	assert.NoError(t, err)
	n = t1.Next()
	assert.Equal(t, int64(1), n)
	err = t1.Done(n)
	assert.NoError(t, err)
	err = t1.Reset(0)
	assert.NoError(t, err)
	err = o.Close()
	assert.NoError(t, err)
	o2, err := os.OpenFile(filepath.Join(dir, "demo.wal"), os.O_RDWR, 0600)
	assert.NoError(t, err)
	t2, err := ReadFromWal(o2, 32)
	assert.NoError(t, err)
	assert.Equal(t, int64(32), t2.size)
	n = t2.Next()
	assert.Equal(t, int64(0), n)
}
