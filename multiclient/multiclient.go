package multiclient

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Multiclient struct {
	buffer  int
	clients map[*url.URL]*http.Client
}

func New(buffer int, reqs ...*http.Request) (*Multiclient, error) {
	mc := &Multiclient{
		buffer:  buffer,
		clients: make(map[*url.URL]*http.Client, len(reqs)),
	}
	for _, req := range reqs {
		if req.Method != http.MethodGet {
			return nil, fmt.Errorf("Only GET method is handled, not %s", req.Method)
		}
		mc.clients[req.URL] = &http.Client{}
		mc.clients[req.URL].Timeout = 3 * time.Second
	}
	w := &sync.WaitGroup{}
	w.Add(len(reqs))
	s := &sync.Mutex{}
	broken := make([]*url.URL, 0)
	size := 0
	for _, req := range reqs {
		req.Method = http.MethodHead
		go func() {
			resp, err := mc.clients[req.URL].Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Println(err)
				s.Lock()
				broken = append(broken, req.URL)
				s.Unlock()
			} else {
				cl, err := strconv.Atoi(resp.Header.Get("content-length"))
				if err != nil {
					log.Println(err)
					s.Lock()
					broken = append(broken, req.URL)
					s.Unlock()
				} else {
					if size == 0 {
						size = cl
					} else {
						if size != cl {
							panic("Different size")
						}
					}
				}
			}
			w.Done()
		}()
	}
	w.Wait()
	for _, req := range reqs {
		go func() {
			resp, err := c.Do(req)
		}()
	}
	return mc, nil
}

func (*Multiclient) Download() error {

}
