package multiclient

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Cake struct {
	lock *sync.Mutex
	file io.WriteSeeker
}

func NewCake(file io.WriteSeeker) *Cake {
	return &Cake{
		lock: &sync.Mutex{},
		file: file,
	}
}

func (c *Cake) Bite(offset int64, body io.Reader, size int64) error {
	buff, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	_, err = c.file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	n, err := c.file.Write(buff)
	if err != nil {
		return err
	}
	if int64(n) != size {
		return fmt.Errorf("Uncomplete write %d of %d", n, size)
	}
	syncer, ok := c.file.(*os.File)
	if ok {
		err = syncer.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}
