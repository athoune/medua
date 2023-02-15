package multiclient

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	dir, err := os.MkdirTemp("", "download")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up
	file := filepath.Join(dir, "demo.iso")
	f, err := os.OpenFile(file, os.O_CREATE+os.O_WRONLY, 0644)
	assert.NoError(t, err)
	h := sha256.New()
	o, err := os.Open("/dev/random")
	assert.NoError(t, err)
	noise := make([]byte, 5*1024)
	for i := 0; i < 1024; i++ {
		n, err := o.Read(noise)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, 5*1024, n)
		_, err = f.Write(noise[:n])
		if err != nil {
			panic(err)
		}
		h.Write(noise)
	}
	s, err := f.Stat()
	assert.NoError(t, err)
	fmt.Println("file size", s.Size())
	f.Close()
	ts1 := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer ts1.Close()
	ts2 := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer ts2.Close()

	reqs := make([]*http.Request, 2)
	for i := 0; i < 2; i++ {
		reqs[i], err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/demo.iso", []string{ts1.URL, ts2.URL}[i]), nil)
		assert.NoError(t, err)
	}

	client := New(1024 * 1024)

	testPath := filepath.Join(dir, "out")
	out, err := os.OpenFile(testPath, os.O_WRONLY+os.O_CREATE, 0644)
	assert.NoError(t, err)
	err = client.Download(out, nil, reqs...)
	assert.NoError(t, err)
	err = out.Close()
	assert.NoError(t, err)
	s2, err := os.Stat(testPath)
	assert.NoError(t, err)

	assert.Equal(t, s.Size(), s2.Size(), "input and output has same size")
	out, err = os.Open(testPath)
	assert.NoError(t, err)
	defer out.Close()
	h2 := sha256.New()
	_, err = io.Copy(h2, out)
	assert.NoError(t, err)
	assert.Equal(t, h.Sum(nil), h2.Sum(nil))

}
