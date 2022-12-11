package multiclient

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
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

func (mc *Multiclient) Download(writer io.Writer, reqs ...*http.Request) error {
	for _, req := range reqs {
		if req.Method != http.MethodGet {
			return fmt.Errorf("Only GET method is handled, not %s", req.Method)
		}
		_, ok := mc.clients[req.URL.Host]
		if !ok {
			mc.clients[req.URL.Host] = &http.Client{
				Timeout: 3 * time.Second,
			}
		}
	}
	w := &sync.WaitGroup{}
	w.Add(len(reqs))
	size := 0
	//range := 0
	s := &sync.Mutex{}
	usable := make([]*http.Request, 0)
	for _, req := range reqs {
		req.Method = http.MethodHead
		go func() {
			defer w.Done()
			resp, err := mc.clients[req.URL.Host].Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Println(err)
				return
			}
			cl, err := strconv.Atoi(resp.Header.Get("content-length"))
			if err != nil {
				log.Println(err)
				return
			}
			if size == 0 {
				size = cl
			} else {
				if size != cl {
					panic("Different size")
				}
			}
			if resp.Header.Get("Accept-Range") != "bytes" {
				log.Println("Accept-Range is mandatory")
				return
			}
			s.Lock()
			usable = append(usable, req)
			s.Unlock()
		}()
	}
	w.Wait()

	w.Add(len(usable))
	for _, req := range usable {
		go func() {
			defer w.Done()
			resp, err := mc.clients[req.URL.Host].Do(req)
			if err != nil {
				log.Fatal(err)
				return
			}
			if resp.StatusCode != http.StatusOK {

			}
		}()
	}
	return nil
}
