package todo

import (
	"encoding/binary"
	"io"
	"os"
)

type Line struct {
	Doing bool
	Chunk int64
}

type Wal struct {
	file *os.File
	size int64
}

func CreateWal(file *os.File, size int64) (*Wal, error) {
	err := binary.Write(file, binary.LittleEndian, size)
	if err != nil {
		return nil, err
	}
	err = file.Sync()
	if err != nil {
		return nil, err
	}
	return &Wal{
		file: file,
		size: size,
	}, nil
}

func ReadWal(file *os.File) (*Todo, error) {
	var size int64
	err := binary.Read(file, binary.LittleEndian, &size)
	if err != nil {
		return nil, err
	}
	todo := New(size)
	todo.wal = &Wal{
		file: file,
		size: size,
	}
	var line Line
	for {
		err = binary.Read(file, binary.LittleEndian, &line)
		if err != nil {
			if err == io.EOF { // wal reading is complete
				break
			}
			return nil, err
		}
		// do or undo ?
		todo.doing[line.Chunk] = line.Doing
	}
	// cursor is the first false in the list
	for n, doing := range todo.doing {
		if !doing {
			todo.cursor = int64(n)
			break
		}
	}

	return todo, nil
}

func (w *Wal) Done(chunk int64) error {
	return w.log(Line{
		Doing: true,
		Chunk: chunk,
	})
}

func (w *Wal) Undo(chunk int64) error {
	return w.log(Line{
		Doing: false,
		Chunk: chunk,
	})
}

func (w *Wal) log(chunk Line) error {
	err := binary.Write(w.file, binary.LittleEndian, chunk)
	if err != nil {
		return err
	}
	return w.file.Sync()
}
