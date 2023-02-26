package cake

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCake(t *testing.T) {
	dir, err := os.MkdirTemp("", "download")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up
	file := filepath.Join(dir, "demo.iso")
	f, err := os.OpenFile(file, os.O_CREATE+os.O_RDWR, 0644)
	assert.NoError(t, err)

	cake := New(f)
	err = cake.Bite(3, bytes.NewBuffer([]byte("lo")), 2)
	assert.NoError(t, err)
	err = cake.Bite(0, bytes.NewBuffer([]byte("hel")), 3)
	assert.NoError(t, err)

	err = f.Sync()
	assert.NoError(t, err)

	_, err = f.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	content, err := io.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(content))
}
