package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/athoune/medusa/multiclient"
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
	err = mc.Download(dest, wal, urls...)
	if err != nil {
		log.Fatal(err)
	}
}
