package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/athoune/medusa/multiclient"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Give me some urls")
	}
	urls := make([]*http.Request, len(os.Args)-1)
	for i, u := range os.Args[1:] {
		uu, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		urls[i] = &http.Request{
			Method: http.MethodGet,
			URL:    uu,
		}
	}
	mc := multiclient.New(10 * 1024 * 1024)
	tmp, err := os.Open("/tmp/medusa")
	if err != nil {
		log.Fatal(err)
	}
	err = mc.Download(tmp, urls...)
	if err != nil {
		log.Fatal(err)
	}
	tmp.Close()
}
