package scroll_test

import (
	"encoding/gob"
	"fmt"
	"github.com/uchicago-sg/scroll"
	"io/ioutil"
	"os"
)

func read(c scroll.Cursor) {
	var x Once
	err := c.Next(&x)
	if err != nil {
		fmt.Printf("err: %s\n", err)
	} else {
		fmt.Printf("read: %s\n", x)
	}
}

type Once string

func (o Once) Key() string { return string(o) }

func ExampleMemoryLog() {
	log := scroll.MemoryLog()
	cursor := log.Cursor()

	read(cursor)
	log.Append(Once("peach"))
	read(cursor)

	log.Append(Once("banana"))
	log.Append(Once("pear"))
	log.Append(Once("banana"))
	read(cursor)
	read(cursor)
	read(cursor)

	// Output:
	// err: scroll.Done: no more entries
	// read: peach
	// read: pear
	// read: banana
	// err: scroll.Done: no more entries
}

func ExampleFileLog() {
	gob.Register(Once(""))

	file, _ := ioutil.TempFile("", "")
	defer os.Remove(file.Name())

	log := scroll.FileLog(file.Name())
	cursor := log.Cursor()

	read(cursor)
	log.Append(Once("peach"))
	read(cursor)

	log.Append(Once("banana"))
	log.Append(Once("pear"))
	log.Append(Once("banana"))
	read(cursor)
	read(cursor)
	read(cursor)

	// Output:
	// err: scroll.Done: no more entries
	// read: peach
	// read: pear
	// read: banana
	// err: scroll.Done: no more entries
}
