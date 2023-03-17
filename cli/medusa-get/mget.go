package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/athoune/medusa/multiclient"
	"github.com/athoune/medusa/widgets"
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
	mc.Timeout = 15 * time.Second

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
	head.SetBorder(true)
	head.Write([]byte("Medusa rulez\n"))
	grid.AddItem(head, 5, 1, false).SetDirection(tview.FlexRow)

	log.SetOutput(head)
	d := mc.Download(dest, wal, urls...)
	downloaders := make([]string, 0)
	d.OnHead = func(_head multiclient.Head) {
		downloaders = append(downloaders, _head.Domain)
		app.QueueUpdateDraw(func() {
			head.SetLabel(_head.Domain)
		})
	}
	d.OnHeadEnd = func() {
		app.QueueUpdateDraw(func() {
			// body.Clear()
			log.SetOutput(head)
			tiles = widgets.NewTiles(int(d.ContentLength))
			tiles.AddHosts(downloaders...)
			tiles.SetBorder(true)
			grid.AddItem(tiles, len(downloaders)+2, 1, true).SetDirection(tview.FlexRow)
			head.Clear()
		})
	}
	d.OnChunk = func(chunk multiclient.Chunk) {
		app.QueueUpdateDraw(func() {
			tiles.AckChunk(chunk.Name)
			app.Sync()
		})
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
}
