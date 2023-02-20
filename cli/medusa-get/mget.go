package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/athoune/medusa/multiclient"
	"github.com/cheggaaa/pb"
)

func main() {
	if len(os.Args) <= 2 {
		log.Fatal(os.Args[0], "destination urls…")
	}
	urls := make([]*http.Request, len(os.Args)-2)
	for i, u := range os.Args[2:] {
		uu, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		urls[i] = &http.Request{
			Method: http.MethodGet,
			URL:    uu,
		}
	}
	wal, err := os.OpenFile(os.Args[1]+".wal", os.O_CREATE+os.O_RDWR, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer wal.Close()
	mc := multiclient.New(1024 * 1024) // 1Mo
	mc.Timeout = 15 * time.Second

	dest, err := os.OpenFile(os.Args[1], os.O_WRONLY+os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer dest.Close()

	bars := make([]*pb.ProgressBar, 0)
	pbs := make(map[string]*pb.ProgressBar)
	maxSize := 0
	onHead := func(head multiclient.Head) {
		if len(head.Domain) > maxSize {
			maxSize = len(head.Domain)
		}
		_, ok := pbs[head.Domain]
		if !ok {
			bar := pb.New(int(head.Size))
			bar.ShowPercent = false
			bar.ShowSpeed = true
			bars = append(bars, bar)
			pbs[head.Domain] = bar
		}
	}
	var pool *pb.Pool
	onHeadEnd := func() {
		namePadding := fmt.Sprintf("%%-%ds", maxSize)
		for name, bar := range pbs {
			pbs[name] = bar.Prefix(fmt.Sprintf(namePadding, name))
		}
		var err error
		pool, err = pb.StartPool(bars...)
		if err != nil {
			log.Fatal(err)
		}
	}
	onChunk := func(chunk multiclient.Chunk) {
		pbs[chunk.Name].Set(chunk.Count).Finish()
	}
	err = mc.Download(dest, wal, onHead, onHeadEnd, onChunk, urls...)
	if err != nil {
		log.Fatal(err)
	}
	pool.Stop()
}
