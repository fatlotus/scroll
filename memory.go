package scroll

import (
	"golang.org/x/net/context"
	"reflect"
	"sync"
)

type memLog struct {
	Objects   []interface{}
	LastIndex map[string]int
	sync.Mutex
}

type memCursor struct {
	Log    *memLog
	Offset int
}

func MemoryLog() Log {
	return &memLog{
		Objects:   make([]interface{}, 0),
		LastIndex: make(map[string]int),
	}
}

func (m *memLog) Cursor() Cursor {
	return &memCursor{m, 0}
}

func (m *memLog) SetContext(c context.Context) error {
	return nil
}

func (c *memCursor) Next(x interface{}) error {
	m := c.Log
	vx := reflect.ValueOf(x)
	m.Lock()
	defer m.Unlock()

	for c.Offset < len(m.Objects) {
		if y, ok := m.Objects[c.Offset].(Unique); ok {
			if m.LastIndex[y.Key()] != c.Offset {
				c.Offset += 1
				continue
			}
		}
		vx.Elem().Set(reflect.ValueOf(m.Objects[c.Offset]))
		c.Offset += 1
		return nil
	}

	return Done
}

func (m *memLog) Append(x interface{}) error {
	m.Lock()
	defer m.Unlock()

	if uniq, ok := x.(Unique); ok {
		m.LastIndex[uniq.Key()] = len(m.Objects)
	}
	m.Objects = append(m.Objects, x)
	return nil
}
