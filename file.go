package scroll

import (
	"encoding/gob"
	"golang.org/x/net/context"
	"io"
	"os"
	"sync"
)

type fileLog struct {
	Path        string
	LastIndex   map[string]int
	File        *os.File
	Encoder     *gob.Encoder
	ObjectCount int
	sync.Mutex
}

type fileCursor struct {
	Log     *fileLog
	Decoder *gob.Decoder
	Offset  int
}

func FileLog(path string) Log {
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	log := &fileLog{
		Path:      path,
		LastIndex: make(map[string]int),
		File:      fp,
		Encoder:   gob.NewEncoder(fp),
	}

	ctx := context.Background()

	var record interface{}
	cursor := log.Cursor().(*fileCursor)
	for {
		if err := cursor.Next(ctx, &record); err != nil {
			if err == Done {
				break
			}
			panic(err)
		}

		if uniq, ok := record.(Unique); ok {
			log.LastIndex[uniq.Key()] = cursor.Offset
		}
		log.ObjectCount += 1
	}

	return log
}

func (l *fileLog) Cursor() Cursor {
	fp, err := os.Open(l.Path)
	if err != nil {
		panic(err)
	}
	return &fileCursor{l, gob.NewDecoder(fp), 0}
}

func (c *fileCursor) Next(ctx context.Context, x interface{}) error {
	l := c.Log

	l.Lock()
	defer l.Unlock()

	for {
		if err := c.Decoder.Decode(x); err != nil {
			if err == io.EOF {
				return Done
			}
			return err
		}
		if y, ok := x.(Unique); ok {
			if l.LastIndex[y.Key()] != c.Offset {
				c.Offset += 1
				continue
			}
		}
		c.Offset += 1
		return nil
	}

	return Done
}

func (l *fileLog) Append(c context.Context, record interface{}) error {
	l.Lock()
	defer l.Unlock()

	if uniq, ok := record.(Unique); ok {
		l.LastIndex[uniq.Key()] = l.ObjectCount
	}

	l.ObjectCount += 1
	err := l.Encoder.Encode(record)
	if err != nil {
		panic(err)
		return err
	}
	err = l.File.Sync()
	if err != nil {
		panic(err)
	}
	return err
}
