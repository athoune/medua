package multiclient

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	wal           *os.File
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
	var todo *_todo.Todo
	if d.wal == nil {
		todo = _todo.New(bites)
	} else {
		var err error
		todo, err = _todo.ReadFromWal(d.wal, bites)
		if err != nil {
			return err
		}
	}

	oops := make(chan error, len(d.reqs))
	for _, req := range d.reqs {
		go func(req *http.Request) {
			name := req.URL.Hostname()
			for {
				b := todo.Next()
				if b == -1 {
					oops <- io.EOF
					return
				}
				err := d.getOne(b*d.biteSize, name, req)
				if err != nil {
					// the fetch has failed, lets retry with another worker
					todo.Reset(b)
					oops <- err
					log.Println("lets stop ", name, err)
					return // lets kill this worker
				}
				err = todo.Done(b) // ack
				if err != nil {
					log.Println("can't write wal", err)
					oops <- err
					return
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
			if workers == 0 {
				if err == io.EOF {
					return nil
				}
				return errors.New("all workers have failed")
			}
		} else {
			bites -= 1
			if bites == 0 {
				return nil
			}
		}
	}
	return nil
}

func (d *Download) getOne(offset int64, name string, r *http.Request) error {
	ts := time.Now()
	end := offset + d.biteSize - 1
	if end >= d.contentLength {
		end = d.contentLength - 1
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
	return nil
}
