package multiclient

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Multiclient struct {
	biteSize int64
	lock     *sync.Mutex
	clients  map[string]*http.Client
	Timeout  time.Duration
}

func New(biteSize int64) *Multiclient {

	return &Multiclient{
		biteSize: biteSize,
		lock:     &sync.Mutex{},
		clients:  make(map[string]*http.Client, 0),
		Timeout:  30 * time.Second,
	}

}

func (mc *Multiclient) LazyClient(host string) *http.Client {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	c, ok := mc.clients[host]
	if !ok {
		c = &http.Client{
			Timeout: mc.Timeout,
		}
		mc.clients[host] = c
	}
	return c
}

func (mc *Multiclient) Download(writer io.WriteSeeker, wal *os.File, reqs ...*http.Request) error {
	for _, req := range reqs {
		if req.Method != http.MethodGet {
			return fmt.Errorf("Only GET method is handled, not %s", req.Method)
		}
	}
	download := &Download{
		reqs:     reqs,
		clients:  mc,
		biteSize: mc.biteSize,
		cake:     NewCake(writer),
		wal:      wal,
	}

	err := download.head()
	if err != nil {
		return err
	}

	err = download.getAll()
	if err != nil {
		return err
	}

	return nil
}
