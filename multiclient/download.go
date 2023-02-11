package multiclient

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Download struct {
	cake          *Cake
	reqs          []*http.Request
	contentLength int64
	wait          *sync.WaitGroup
	lock          *sync.Mutex
	clients       ClientPool
	biteSize      int64
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
	d.wait.Add(len(d.reqs))
	usable := make([]*http.Request, 0)
	crange := 0
	lock := &sync.Mutex{}
	for _, req := range d.reqs {
		req.Method = http.MethodHead
		go func(r *http.Request) {
			ts := time.Now()
			defer d.wait.Done()
			resp, err := d.clients.LazyClient(r.URL.Host).Do(r)
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Println(r.URL, resp.Status, err)
				return
			}
			cl, err := strconv.Atoi(resp.Header.Get("content-length"))
			if err != nil {
				log.Println("Can't parse content-length", err)
				return
			}
			lock.Lock()
			if d.contentLength == -1 {
				d.contentLength = int64(cl)
				lock.Unlock()
			} else {
				defer lock.Unlock()
				if d.contentLength != int64(cl) {
					log.Fatal("Different size ", d.contentLength, cl)
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
			if resp.Header.Get("Accept-Ranges") != "bytes" {
				log.Fatal("Accept-Ranges is mandatory ", resp.Header.Get("Accept-Ranges"))
			}
			// Lets restore initial method : GET
			r.Method = http.MethodGet
			d.lock.Lock()
			usable = append(usable, r)
			d.lock.Unlock()
			log.Println("HEAD latency", r.URL.Host, time.Since(ts))
		}(req)
	}
	d.wait.Wait()
	d.reqs = usable
	return nil
}

func (d *Download) getAll() error {
	bites := d.contentLength / d.biteSize
	if d.contentLength%d.biteSize > 0 {
		bites += 1
	}
	todo := make(chan int64, bites)
	var i int64
	for i = 0; i < bites; i++ {
		todo <- i * d.biteSize
	}

	oops := make(chan error, len(d.reqs))
	for _, req := range d.reqs {
		go func(r *http.Request) {
			name := r.URL.Hostname()
			for b := range todo {
				err := d.getOne(b, name, todo, r)
				if err != nil {
					if err != io.EOF {
						// the fetch has failed, lets retry with another worker
						todo <- b
					}
					oops <- err
					log.Println("lets stop ", name, err)
					return // lets kill this worker
				}
				oops <- nil // one bite done
			}
		}(req)
	}
	var err error
	workers := len(d.reqs)
	for err = range oops {
		if err != nil {
			workers -= 1
			log.Println("Available workers", workers)
			if workers == 0 && err != io.EOF {
				return errors.New("All workers have failed.")
			}
		}
		if err == nil || err == io.EOF {
			bites -= 1
			if bites == 0 {
				return nil
			}
		} else {
			log.Println(err)
		}
	}
	return nil
}

func (d *Download) getOne(offset int64, name string, todo chan int64, r *http.Request) error {
	ts := time.Now()
	eof := false
	end := offset + d.biteSize - 1
	if end >= d.contentLength {
		end = d.contentLength - 1
		eof = true
	}
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	r.Header.Set("user-agent", "Medusa")
	r.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, end))
	resp, err := d.clients.LazyClient(r.URL.Host).Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		log.Println("Can't fetch ", resp.Status, r.Header)
		return fmt.Errorf("Bad status %s", resp.Status)
	}
	defer resp.Body.Close()
	err = d.cake.Bite(offset, resp.Body, end-offset+1)
	if err != nil {
		log.Printf("%s can't write %d-%d content length: %s err: %s\n",
			name, offset, end,
			resp.Header.Get("content-length"), err)
		return err
	}
	log.Printf("%s %d-%d %v\n", r.URL.Host, offset, end, time.Since(ts))
	if eof {
		return io.EOF
	}
	return nil
}
