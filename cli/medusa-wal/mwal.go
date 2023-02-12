package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/athoune/medusa/todo"
)

func main() {
	o, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	for {
		var line todo.Line
		err = binary.Read(o, binary.LittleEndian, &line)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Printf("%v %v\n", line.Doing, line.Chunk)
	}

}
