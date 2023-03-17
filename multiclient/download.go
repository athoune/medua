package multiclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/athoune/medusa/cake"
	_todo "github.com/athoune/medusa/todo"
)

type Chunk struct {
	Name string
	Size int
}

type Head struct {
	Domain  string
	Latency time.Duration
	Size    int64
}

type Download struct {
	cake          *cake.Cake
	reqs          []*http.Request
	ContentLength int64
	lock          *sync.Mutex
	client        *http.Client
	biteSize      int64
	written       int
	wal           *os.File
	OnHead        func(Head)
	OnHeadEnd     func()
	OnChunk       func(Chunk)
}

func (d *Download) clean() {
	d.ContentLength = -1
	d.lock = &sync.Mutex{}
	d.written = 0
}

func (d *Download) Fetch() error {
	err := d.preflight()
	if err != nil {
		return err
	}
	err = d.head()
	if err != nil {
		return err
	}
	return d.getAll()
}

func (d *Download) preflight() error {
	for _, req := range d.reqs {
		if req.Method != http.MethodGet {
			return fmt.Errorf("only GET method is handled, not %s", req.Method)
		}
		ips, err := net.LookupIP(strings.Split(req.URL.Host, ":")[0])
		if err != nil {
			return err
		}
		log.Println(req.URL.Host, ips)
	}
	return nil
}

func (d *Download) getAll() error {
	multi := 3
	bites := d.ContentLength / d.biteSize
	if d.ContentLength%d.biteSize > 0 {
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
		for i := 0; i < multi; i++ {
			go func(req *http.Request, i int) {
				name := fmt.Sprintf("%s#%d", req.URL.Hostname(), i)
				cpt := 0
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
					cpt++
					if d.OnChunk != nil {
						d.OnChunk(Chunk{
							Name: name,
							Size: cpt * int(d.biteSize),
						})
					}
					oops <- nil // one bite done
				}
			}(req.Clone(context.TODO()), i)
		}
	}
	var err error
	workers := len(d.reqs) * multi
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
	//ts := time.Now()
	end := offset + d.biteSize - 1
	if end >= d.ContentLength {
		end = d.ContentLength - 1
	}
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	r.Header.Set("user-agent", "Medusa")
	r.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, end))
	resp, err := d.client.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		log.Println("Can't fetch ", resp.Status, r.Header)
		return fmt.Errorf("bad status %s", resp.Status)
	}
	defer resp.Body.Close()
	err = d.cake.Bite(offset, resp.Body, end-offset+1)
	if err != nil {
		log.Printf("%s can't write %d-%d content length: %s err: %s\n",
			name, offset, end,
			resp.Header.Get("content-length"), err)
		return err
	}
	//log.Printf("%s %d-%d %v\n", r.URL.Host, offset, end, time.Since(ts))
	return nil
}
