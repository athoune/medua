package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/athoune/medusa/multiclient"
	"github.com/athoune/medusa/widgets"
	"github.com/docker/go-units"
	"github.com/rivo/tview"
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
	mc.Timeout = 5 * time.Second

	dest, err := os.OpenFile(os.Args[1], os.O_WRONLY+os.O_CREATE, 0600)
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
			dbt := float64(time.Second) * float64(d.Written()) / float64(time.Since(start))
			ratio := 100 * float64(d.Written()) / float64(d.ContentLength)
			footer.Clear()
			fmt.Fprintf(footer, "%02d%% %s/s\n", int64(ratio), units.HumanSize(dbt))
			app.Sync()
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
