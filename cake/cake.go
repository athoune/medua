package cake

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Cake can be eaten bite by bite, even unordered
type Cake struct {
	lock *sync.Mutex
	file io.WriteSeeker
}

func New(file io.WriteSeeker) *Cake {
	return &Cake{
		lock: &sync.Mutex{},
		file: file,
	}
}

// Bite set the cursor, read the body then copy expected size.
// Thread safe.
func (c *Cake) Bite(offset int64, body io.Reader, size int64) error {
	buff, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if int64(len(buff)) != size {
		return fmt.Errorf("uncomplete read %d of %d", len(buff), size)
	}
	// read can be slow and async, lock only the write step
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
		return fmt.Errorf("uncomplete write %d of %d", n, size)
	}
	syncer, ok := c.file.(*os.File) // if it's a File, lets sync
	if ok {
		err = syncer.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}
