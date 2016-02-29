package scroll

import (
	"errors"
	"golang.org/x/net/context"
)

var Done = errors.New("scroll.Done: no more entries")

type Log interface {
	Cursor() Cursor
	Append(x interface{}) error
	SetContext(c context.Context) error
}

type Cursor interface {
	Next(x interface{}) error
}

type Unique interface {
	Key() string
}
