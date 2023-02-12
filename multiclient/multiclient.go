package multiclient

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Multiclient struct {
	biteSize int64
	lock     *sync.Mutex
	client   *http.Client
	Timeout  time.Duration
}

func New(biteSize int64) *Multiclient {

	return &Multiclient{
		biteSize: biteSize,
		lock:     &sync.Mutex{},
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
		Timeout: 30 * time.Second,
	}

}

func (mc *Multiclient) Download(writer io.WriteSeeker, wal *os.File, reqs ...*http.Request) error {
	for _, req := range reqs {
		if req.Method != http.MethodGet {
			return fmt.Errorf("Only GET method is handled, not %s", req.Method)
		}
		ips, err := net.LookupIP(strings.Split(req.URL.Host, ":")[0])
		if err != nil {
			return err
		}
		log.Println(req.URL.Host, ips)
	}
	download := &Download{
		reqs:     reqs,
		client:   mc.client,
		biteSize: mc.biteSize,
		cake:     NewCake(writer),
		wal:      wal,
	}

	err := download.head()
	if err != nil {
		return err
	}

	err = download.getAll()
	if err != nil {
		return err
	}

	return nil
}
