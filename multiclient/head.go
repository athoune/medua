package multiclient

import (
	"context"
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
			resp, err := d.client.Do(r)
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
			if d.ContentLength == -1 {
				d.ContentLength = int64(cl)
				lock.Unlock()
			} else {
				defer lock.Unlock()
				if d.ContentLength != int64(cl) {
					events <- fmt.Errorf("Different size %d %d", d.ContentLength, cl)
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
			latency := time.Since(ts)
			r = r.WithContext(context.WithValue(r.Context(), "latency", latency))

			d.lock.Lock()
			usable = append(usable, r)
			d.lock.Unlock()

			if d.OnHead != nil {
				for i := 0; i < 3; i++ {
					d.OnHead(Head{
						Domain:  fmt.Sprintf("%s#%d", r.URL.Hostname(), i),
						Latency: latency,
						Size:    d.ContentLength,
					})
				}
			}
			events <- nil
		}(req)
	}
	var err error
	for i := 0; i < len(d.reqs); i++ {
		err = <-events
		if err != nil {
			return err
		}
	}
	if d.OnHeadEnd != nil {
		d.OnHeadEnd()
	}
	d.reqs = usable
	return nil
}
