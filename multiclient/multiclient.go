package multiclient

import (
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Multiclient struct {
	biteSize int64
	lock     *sync.Mutex
	client   *http.Client
	Timeout  time.Duration
}

func New(biteSize int64) *Multiclient {
	return &Multiclient{
		biteSize: biteSize,
		lock:     &sync.Mutex{},
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
		Timeout: 30 * time.Second,
	}

}

func (mc *Multiclient) Download(writer io.WriteSeeker, wal *os.File, reqs ...*http.Request) *Download {
	return NewDownload(mc.client, mc.biteSize, writer, wal, reqs...)
}
