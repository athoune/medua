package multiclient

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Multiclient struct {
	buffer  int
	clients map[string]*http.Client
}

func New(buffer int) *Multiclient {

	return &Multiclient{
		buffer:  buffer,
		clients: make(map[string]*http.Client, 0),
	}

}

type ClientPool interface {
	LazyClient(string) *http.Client
}

func (mc *Multiclient) LazyClient(host string) *http.Client {
	c, ok := mc.clients[host]
	if !ok {
		c = &http.Client{
			Timeout: 3 * time.Second,
		}
		mc.clients[host] = c
	}
	return c
}

func (mc *Multiclient) Download(writer io.Writer, reqs ...*http.Request) error {
	for _, req := range reqs {
		if req.Method != http.MethodGet {
			return fmt.Errorf("Only GET method is handled, not %s", req.Method)
		}
	}
	download := &Download{
		writer:  writer,
		reqs:    reqs,
		clients: mc,
	}

	err := download.head()
	if err != nil {
		return err
	}

	return nil
}
