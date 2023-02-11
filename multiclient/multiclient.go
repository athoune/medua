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
	timeout  time.Duration
}

func New(biteSize int64) *Multiclient {

	return &Multiclient{
		biteSize: biteSize,
		lock:     &sync.Mutex{},
		clients:  make(map[string]*http.Client, 0),
		timeout:  30 * time.Second,
	}

}

type ClientPool interface {
	LazyClient(string) *http.Client
}

func (mc *Multiclient) LazyClient(host string) *http.Client {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	c, ok := mc.clients[host]
	if !ok {
		c = &http.Client{
			Timeout: mc.timeout,
		}
		mc.clients[host] = c
	}
	return c
}

func (mc *Multiclient) Download(writer io.WriteSeeker, reqs ...*http.Request) error {
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
	}

	err := download.head()
	if err != nil {
		return err
	}

	err = download.getAll()
	if err != nil {
		return err
	}

	f, ok := writer.(*os.File)
	if ok {
		err = f.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}
