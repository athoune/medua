package todo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodo(t *testing.T) {
	todo := New(5)
	n, err := todo.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	n, err = todo.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
	n, err = todo.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), n)
	assert.NoError(t, todo.Reset(1))
	n, err = todo.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
	n, err = todo.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), n)
	n, err = todo.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(4), n)
	n, err = todo.Next()
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), n)
	assert.Error(t, todo.Reset(5))
}

func TestAsync(t *testing.T) {
	todo := New(500)
	done := make(chan int64, 10)
	for i := 0; i < 100; i++ {
		go func() {
			for {
				d, err := todo.Next()
				assert.NoError(t, err)
				log.Println(d)
				if d == -1 {
					break
				}
				done <- d
			}
		}()
	}
	kv := make(map[int64]interface{})
	for d := range done {
		kv[d] = new(interface{})
		if len(kv) == 500 {
			break
		}
	}
}
