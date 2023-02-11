package multiclient

import (
	"bytes"
	"sync"
)

type Buffer struct {
	buffer bytes.Buffer
	lock   sync.Mutex
}
