package widgets

import (
	"sync"

	"github.com/athoune/medusa/multiclient"
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
	last          int64
	lock          *sync.Mutex
	display       [4]int
	Stopped       []string
	done          func() []bool
}

func NewTiles(n int, done func() []bool) *Tiles {
	t := &Tiles{
		Box:           tview.NewBox(),
		numberOfChunk: n,
		tiles:         make(map[string][]uint8),
		lock:          &sync.Mutex{},
		display:       [4]int{},
		Stopped:       make([]string, 0),
		done:          done,
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

func (t *Tiles) AckChunk(chunk multiclient.Chunk) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.tiles[chunk.Name][chunk.Poz] = 1
	t.last = chunk.Poz
	t.poz++
}

func (t *Tiles) Draw(screen tcell.Screen) {
	t.lock.Lock()
	defer t.lock.Unlock()
	x, y, width, height := t.GetInnerRect()
	i := 0
	var r rune
	if t.display[0] != x || t.display[1] != y || t.display[2] != width || t.display[3] != height {
		t.display = [4]int{x, y, width, height}
		t.SetBackgroundColor(tcell.ColorBlack)
		t.DrawForSubclass(screen, t)
		for _, k := range t.keys {
			back := tcell.ColorBlack
			if i%2 == 0 {
				back = tcell.ColorDarkSlateGray
			}
			kk := []rune(k)
			for a := 0; a < t.maxSize; a++ {
				if a < len(kk) {
					r = kk[a]
				} else {
					if a < t.maxSize-1 {
						r = '.'
					} else {
						r = ' '
					}
				}
				screen.SetContent(x+a, y+i, r, nil, tcell.StyleDefault.Background(back))
			}
			i++
		}
	}
	done := t.done()
	for i, d := range done {
		if d {
			t.poz = i
		}
	}
	var start int
	barWidth := width - t.maxSize
	if t.poz > barWidth {
		start = t.poz - barWidth
	} else {
		start = 0
	}
	i = 0
	for _, k := range t.keys {
		v := t.tiles[k]
		back := tcell.ColorBlack
		if i%2 == 0 {
			back = tcell.ColorDarkSlateGray
		}
		for j := start; j < t.poz; j++ {
			if v[j] == 1 {
				if int64(j) == t.last {
					r = '📦' // tview.BlockLightShade
				} else {
					r = tview.BlockFullBlock
				}
			} else {
				r = '.'
			}
			screen.SetContent(x+t.maxSize+j-start, y+i, r, nil, tcell.StyleDefault.Background(back))
		}
		for _, name := range t.Stopped {
			if name == k {
				screen.SetContent(x+t.maxSize-1, y+i, '🪦', nil, tcell.StyleDefault.Background(back))
			}
		}
		i++
	}
	for j := start; j < t.poz; j++ {
		cpt := 0
		total := 0
		for i := 0; i < int(t.last); i++ {
			if done[i] {
				cpt++
			}
			total++
		}
		/* FIXME result is weird
		for i, f := range []rune(fmt.Sprintf("%d%%", 100*cpt/total)) {
			screen.SetContent(x+i, len(t.keys)+y, f, nil, tcell.StyleDefault)
		}
		*/
		if j < len(done) {
			if done[j] {
				r = tview.BlockFullBlock
			} else {
				r = tview.BlockLightShade
			}
			screen.SetContent(x+t.maxSize+j-start, len(t.keys)+y, r, nil, tcell.StyleDefault)
		}
	}
}
