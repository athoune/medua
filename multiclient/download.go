package multiclient

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Download struct {
	writer        io.Writer
	reqs          []*http.Request
	contentLength int
	wait          *sync.WaitGroup
	lock          *sync.Mutex
	clients       ClientPool
	biteSize      int
	written       int
}

func (d *Download) clean() {
	d.contentLength = -1
	d.wait = &sync.WaitGroup{}
	d.lock = &sync.Mutex{}
	d.written = 0
}

func (d *Download) head() error {
	d.clean()
	usable := make([]*http.Request, 0)
	crange := 0
	for _, req := range d.reqs {
		req.Method = http.MethodHead
		go func(r *http.Request) {
			defer d.wait.Done()
			resp, err := d.clients.LazyClient(r.URL.Host).Do(r)
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Println(err)
				return
			}
			cl, err := strconv.Atoi(resp.Header.Get("content-length"))
			if err != nil {
				log.Println(err)
				return
			}
			if d.contentLength == 0 {
				d.contentLength = cl
			} else {
				if d.contentLength != cl {
					panic("Different size")
				}
			}
			ra := r.Header.Get("range")
			if ra != "" {
				rav, err := strconv.Atoi(ra)
				if err != nil {
					panic(err)
				}
				if crange == -1 {
					crange = rav
				} else {
					if crange != rav {
						// FIXME
						panic(fmt.Errorf("range mismatch: %d != %d", crange, rav))
					}
				}
			}
			if resp.Header.Get("Accept-Range") != "bytes" {
				log.Println("Accept-Range is mandatory")
				return
			}
			d.lock.Lock()
			usable = append(usable, r)
			d.lock.Unlock()
		}(req)
	}
	d.wait.Wait()
	d.reqs = usable
	return nil
}

func (d *Download) get() error {
	for d.written <= d.contentLength {
		d.wait.Add(len(d.reqs))
		for _, req := range d.reqs {
			go func(r *http.Request) {
				defer d.wait.Done()
				resp, err := d.clients.LazyClient(r.URL.Host).Do(r)
				if err != nil {
					log.Fatal(err)
					return
				}
				if resp.StatusCode != http.StatusOK {

				}
			}(req)
		}
		d.wait.Wait()
	}
	return nil
}
