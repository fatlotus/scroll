package scroll

import (
	"errors"
	"golang.org/x/net/context"
)

// Returned by cursor.Next() when there are no new entries in the list.
// Eventual consistency might mean spurious Done errors.
var Done = errors.New("scroll.Done: no more entries")

// A Log is a persistent, stateful transaction list. It supports reading, by
// way of Cursors, and Appending. Because Logs are intended to cache data
// between requests, there is an option SetContext that allows the use of
// another context object.
type Log interface {
	// Create and return a new Cursor at the beginning of the log.
	Cursor() Cursor

	// Adds a new entry to the end of the log.
	Append(c context.Context, x interface{}) error
}

// A Cursor is a readable stream of entities from a Log. Once the cursor has
// reached the end of the Log, the Next function returns Done.
type Cursor interface {
	// Retrieve the next entry from the log.
	Next(c context.Context, x interface{}) error
}

// A Unique is an object that overwrites previous elements in the log.
// Typically, Unique is used where an entry is an update that completely
// replaces an earlier one.
type Unique interface {
	Key() string
}
