package multiclient

import (
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

func (c *Cake) Bite(offset int64, body io.Reader) error {
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
	_, err = c.file.Write(buff)
	if err != nil {
		return err
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
