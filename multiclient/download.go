package multiclient

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	_todo "github.com/athoune/medusa/todo"
)

type Download struct {
	cake          *Cake
	reqs          []*http.Request
	contentLength int64
	lock          *sync.Mutex
	clients       ClientPool
	biteSize      int64
	written       int
}

func (d *Download) clean() {
	d.contentLength = -1
	d.lock = &sync.Mutex{}
	d.written = 0
}

func (d *Download) getAll() error {
	bites := d.contentLength / d.biteSize
	if d.contentLength%d.biteSize > 0 {
		bites += 1
	}
	todo := _todo.New(bites)

	oops := make(chan error, len(d.reqs))
	for _, req := range d.reqs {
		go func(r *http.Request) {
			name := r.URL.Hostname()
			for {
				b := todo.Next()
				if b == -1 {
					break
				}
				err := d.getOne(b*d.biteSize, name, r)
				if err != nil {
					if err != io.EOF {
						// the fetch has failed, lets retry with another worker
						todo.Reset(b)
					}
					oops <- err
					log.Println("lets stop ", name, err)
					return // lets kill this worker
				}
				err = todo.Done(b)
				if err != nil {
					oops <- err
					break
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

func (d *Download) getOne(offset int64, name string, r *http.Request) error {
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
