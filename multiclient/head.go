package multiclient

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func (d *Download) head() error {
	d.clean()
	usable := make([]*http.Request, 0)
	crange := 0
	lock := &sync.Mutex{}
	events := make(chan error, len(d.reqs))
	for _, req := range d.reqs {
		req.Method = http.MethodHead
		go func(r *http.Request) {
			ts := time.Now()
			resp, err := d.clients.LazyClient(r.URL.Host).Do(r)
			if err != nil || resp.StatusCode != http.StatusOK {
				if resp != nil {
					log.Println(r.URL, resp.Status, err)
				} else {
					log.Println(r.URL, err)
				}
				events <- nil
				return
			}
			if resp.Header.Get("Accept-Ranges") != "bytes" {
				events <- fmt.Errorf("Accept-Ranges is mandatory %s", resp.Header.Get("Accept-Ranges"))
				return
			}
			cl, err := strconv.Atoi(resp.Header.Get("content-length"))
			if err != nil {
				events <- fmt.Errorf("Can't parse content-length %v", err)
				return
			}
			lock.Lock()
			if d.contentLength == -1 {
				d.contentLength = int64(cl)
				lock.Unlock()
			} else {
				defer lock.Unlock()
				if d.contentLength != int64(cl) {
					events <- fmt.Errorf("Different size %d %d", d.contentLength, cl)
					return
				}
			}
			ra := r.Header.Get("range")
			if ra != "" {
				rav, err := strconv.Atoi(ra)
				if err != nil {
					events <- fmt.Errorf("Can't read range %v", err)
					return
				}
				if crange == -1 {
					crange = rav
				} else {
					if crange != rav {
						events <- fmt.Errorf("range mismatch: %d != %d", crange, rav)
						return
					}
				}
			}
			// Lets restore initial method : GET
			r.Method = http.MethodGet

			d.lock.Lock()
			usable = append(usable, r)
			d.lock.Unlock()
			events <- nil
			log.Println("HEAD latency", resp.Proto, r.URL.Host, time.Since(ts))
		}(req)
	}
	var err error
	for i := 0; i < len(d.reqs); i++ {
		err = <-events
		if err != nil {
			return err
		}
	}
	d.reqs = usable
	return nil
}
