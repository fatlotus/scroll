package scroll_test

import (
	"encoding/gob"
	"fmt"
	"github.com/fatlotus/scroll"
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
)

func read(ctx context.Context, c scroll.Cursor) {
	var x Once
	err := c.Next(ctx, &x)
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
	ctx := context.Background()

	read(ctx, cursor)
	log.Append(ctx, Once("peach"))
	read(ctx, cursor)

	log.Append(ctx, Once("banana"))
	log.Append(ctx, Once("pear"))
	log.Append(ctx, Once("banana"))
	read(ctx, cursor)
	read(ctx, cursor)
	read(ctx, cursor)

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
	ctx := context.Background()

	read(ctx, cursor)
	log.Append(ctx, Once("peach"))
	read(ctx, cursor)

	log.Append(ctx, Once("banana"))
	log.Append(ctx, Once("pear"))
	log.Append(ctx, Once("banana"))
	read(ctx, cursor)
	read(ctx, cursor)
	read(ctx, cursor)

	// Output:
	// err: scroll.Done: no more entries
	// read: peach
	// read: pear
	// read: banana
	// err: scroll.Done: no more entries
}
