package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/athoune/medusa/multiclient"
	"github.com/athoune/medusa/widgets"
	"github.com/docker/go-units"
	"github.com/rivo/tview"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal(os.Args[0], "urls…")
	}
	urls := make([]*http.Request, len(os.Args)-1)
	var filename string
	var slugs []string
	for i, u := range os.Args[1:] {
		uu, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		if i == 0 {
			slugs = strings.Split(uu.Path, "/")
			filename = slugs[len(slugs)-1]
		} else {
			s := strings.Split(uu.Path, "/")
			for i, slug := range slugs {
				if s[len(s)-1] == slug {
					uu.Path += "/" + strings.Join(slugs[i+1:], "/")
					break
				}
			}
		}
		urls[i] = &http.Request{
			Method: http.MethodGet,
			URL:    uu,
		}
	}
	wal, err := os.OpenFile(filename+".wal", os.O_CREATE+os.O_RDWR, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer wal.Close()
	mc := multiclient.New(1024 * 1024) // 1Mo
	mc.Timeout = 5 * time.Second

	dest, err := os.OpenFile(filename, os.O_WRONLY+os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer dest.Close()

	grid := tview.NewFlex()
	grid.SetTitle("Medusa")

	var tiles *widgets.Tiles
	app := tview.NewApplication().SetRoot(grid, true).SetFocus(grid)
	head := tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	})
	head.SetBorder(true).SetTitle(" Medusa ")
	head.Write([]byte("Medusa rulez\n"))
	grid.AddItem(head, 5, 1, false).SetDirection(tview.FlexRow)
	log.SetOutput(head)

	footer := tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	})
	footer.SetBorder(true).SetTitle(" Performances ")

	d := mc.Download(dest, wal, urls...)
	downloaders := make([]string, 0)
	d.OnHead = func(_head multiclient.Head) {
		downloaders = append(downloaders, _head.Domain)
		app.QueueUpdateDraw(func() {
			head.SetLabel(_head.Domain)
		})
	}
	var start time.Time
	d.OnHeadEnd = func() {
		app.QueueUpdateDraw(func() {
			// body.Clear()
			log.SetOutput(head)
			tiles = widgets.NewTiles(int(d.ContentLength), d.Done)
			tiles.AddHosts(downloaders...)
			tiles.SetBorder(true).SetTitle(" Downloads ")
			grid.AddItem(tiles, len(downloaders)+3, 1, true).SetDirection(tview.FlexRow)
			head.Clear()
			grid.AddItem(footer, 3, 1, false).SetDirection(tview.FlexRow)
		})
		start = time.Now()
	}
	d.OnChunk = func(chunk multiclient.Chunk) {
		app.QueueUpdate(func() {
			tiles.AckChunk(chunk)
			dbt := d.Speed()
			ratio := 100 * float64(d.Written()) / float64(d.ContentLength)
			footer.Clear()
			fmt.Fprintf(footer, "%02d%% %s/s ETA: %v\n", int64(ratio), units.HumanSize(dbt), time.Duration(float64(d.ContentLength-d.Written())/dbt*1000000000))
		})
	}
	d.OnStopped = func(name string) {
		tiles.Stopped = append(tiles.Stopped, name)
	}

	go func() {
		err = d.Fetch()
		if err != nil {
			log.Fatal(err)
		}
		app.Stop()
	}()

	err = app.Run()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "Download in %v\n", time.Since(start))
}
