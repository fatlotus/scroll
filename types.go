package scroll

import (
	"errors"
	"golang.org/x/net/context"
)

var Done = errors.New("scroll.Done: no more entries")

// A Log is a persistent, stateful transaction list. It supports reading, by
// way of Cursors, and Appending. Because Logs are intended to cache data
// between requests, there is an option SetContext that allows the use of
// another context object.
type Log interface {
	Cursor() Cursor
	Append(c context.Context, x interface{}) error
}

// A Cursor is a readable stream of entities from a Log. Once the cursor has
// reached the end of the Log, the Next function returns Done.
type Cursor interface {
	Next(c context.Context, x interface{}) error
}

// A Unique is an object that overwrites previous elements in the log.
type Unique interface {
	Key() string
}
