package widgets

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Tiles struct {
	*tview.Box
	numberOfChunk int
	tiles         map[string][]uint8
	keys          []string
	maxSize       int
	poz           int
	lock          *sync.Mutex
}

func NewTiles(n int) *Tiles {
	t := &Tiles{
		Box:           tview.NewBox(),
		numberOfChunk: n,
		tiles:         make(map[string][]uint8),
		lock:          &sync.Mutex{},
	}
	return t
}

func (t *Tiles) AddHosts(hosts ...string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, host := range hosts {
		t.tiles[host] = make([]uint8, t.numberOfChunk)
		if len(host)+1 > t.maxSize {
			t.maxSize = len(host) + 1
		}
	}
	t.keys = hosts
}

func (t *Tiles) AckChunk(host string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.tiles[host][t.poz] = 1
	t.poz++
}

func (t *Tiles) Draw(screen tcell.Screen) {
	t.SetBackgroundColor(tcell.ColorBlack)
	t.DrawForSubclass(screen, t)
	x, y, width, _ := t.GetInnerRect()
	t.lock.Lock()
	defer t.lock.Unlock()
	var start int
	if t.poz > (width - 2 - t.maxSize) {
		start = t.poz - width - 2 - t.maxSize
	} else {
		start = 0
	}
	i := 0
	var r rune
	fullline := make([]rune, width-x)
	for _, k := range t.keys {
		v := t.tiles[k]
		for a := 0; a < width-x; a++ {
			if a < t.maxSize-1 {
				fullline[a] = '.'
			} else {
				fullline[a] = ' '
			}
		}
		copy(fullline, []rune(k))
		for j := start; j < t.poz; j++ {
			if v[j] == 1 {
				if j == t.poz-1 {
					r = '📦' // tview.BlockLightShade
				} else {
					r = tview.BlockFullBlock
				}
			} else {
				r = ' '
			}
			fullline[j-start+t.maxSize] = r
		}
		back := tcell.ColorBlack
		if i%2 == 0 {
			back = tcell.ColorDarkSlateGray
		}
		screen.SetContent(x, y+i, fullline[0], fullline[1:], tcell.StyleDefault.Background(back))
		i++
	}
}
