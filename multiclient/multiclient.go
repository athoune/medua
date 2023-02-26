package multiclient

import (
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/athoune/medusa/cake"
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
	return &Download{
		reqs:     reqs,
		client:   mc.client,
		biteSize: mc.biteSize,
		cake:     cake.New(writer),
		wal:      wal,
	}
}
