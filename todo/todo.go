package todo

import (
	"errors"
	"sync"
)

type Todo struct {
	doing  []bool
	cursor int64
	lock   sync.Mutex
	size   int64
}

func New(size int64) *Todo {
	return &Todo{
		doing:  make([]bool, size),
		cursor: 0,
		size:   size,
	}
}
func (t *Todo) Reset(poz int64) error {
	if poz >= t.size {
		return errors.New("out of bound")
	}
	t.lock.Lock()
	t.doing[poz] = false
	if poz < t.cursor {
		t.cursor = poz
	}
	t.lock.Unlock()
	return nil
}

func (t *Todo) Next() int64 {
	t.lock.Lock()
	defer t.lock.Unlock()
	for i := t.cursor; i < t.size; i++ {
		if !t.doing[i] {
			t.doing[i] = true
			t.cursor = i + 1
			return i
		}
	}
	return -1
}
