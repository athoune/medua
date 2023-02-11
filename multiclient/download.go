package multiclient

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
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
	for i, req := range d.reqs {
		go func(name int, r *http.Request) {
			for {
				select {
				case b := <-todo:
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
			}
		}(i, req)
	}
	var err error
	workers := len(d.reqs)
	for err = range oops {
		if err != nil {
			workers -= 1
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
		log.Println("todo", bites)
	}
	return nil
}

func (d *Download) getOne(offset int64, name int, todo chan int64, r *http.Request) error {
	eof := false
	end := offset + d.biteSize - 1
	if end >= d.contentLength {
		end = d.contentLength - 1
		eof = true
	}
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
		log.Printf("%d can't write %d-%d content length: %s err: %s\n",
			name, offset, end,
			resp.Header.Get("content-length"), err)
		return err
	}
	if eof {
		return io.EOF
	}
	return nil
}
