package scroll

import (
	"bytes"
	"encoding/gob"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"sync"
	"time"
)

type operation struct {
	Data  []datastore.ByteString
	Order time.Time
}

type dbLog struct {
	Entity  string
	Context context.Context
	sync.Mutex
}

type dbCursor struct {
	LastOp      *operation
	LastKey     *datastore.Key
	Log         *dbLog
	Pending     []operation
	PendingKeys []*datastore.Key
}

// A DatastoreLog is a Log stored on the Google App Engine Datastore, a
// wrapper around BigTable.
//
//  Cost of log.Append: ~(2 + 0.001 * size of gob payload in bytes) writes.
//  Cost of cursor.Next: ~(0.001) reads.
func DatastoreLog(entity string) Log {
	return &dbLog{
		Entity: entity,
	}
}

func (m *dbLog) Cursor() Cursor {
	return &dbCursor{
		Log: m,
	}
}

func (c *dbCursor) fetchMore(ctx context.Context) (err error) {
	if len(c.Pending) > 0 {
		return
	}

	a := datastore.NewKey(ctx, c.Log.Entity, "root", 0, nil)
	q := datastore.NewQuery(c.Log.Entity).Order("Order").Ancestor(a)
	if c.LastOp != nil {
		q = q.Filter("Order >", c.LastOp.Order)
	}

	c.PendingKeys, err = q.Limit(1000).GetAll(ctx, &c.Pending)
	if len(c.Pending) == 0 {
		return Done
	}
	return
}

func (c *dbCursor) skipDups() {
	for i, _ := range c.Pending {
		if !c.PendingKeys[i].Equal(c.LastKey) {
			c.Pending = c.Pending[i:]
			c.PendingKeys = c.PendingKeys[i:]
			return
		}
	}

	c.PendingKeys = c.PendingKeys[:0]
	c.Pending = c.Pending[:0]
}

func (c *dbCursor) Next(ctx context.Context, x interface{}) error {
	c.Log.Lock()
	defer c.Log.Unlock()

	for len(c.Pending) == 0 {
		if err := c.fetchMore(ctx); err != nil {
			return err
		}
		c.skipDups()
	}

	// ensure we skip this key next time
	c.LastOp, c.Pending = &c.Pending[0], c.Pending[1:]
	c.LastKey, c.PendingKeys = c.PendingKeys[0], c.PendingKeys[1:]

	// process the record contents
	buf := &bytes.Buffer{}
	for _, chunk := range c.LastOp.Data {
		buf.Write(chunk)
	}
	err := gob.NewDecoder(buf).Decode(x)
	if err != nil {
		panic(err)
	}
	return err
}

func (m *dbLog) Append(ctx context.Context, x interface{}) error {
	m.Lock()
	defer m.Unlock()

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(x); err != nil {
		return err
	}

	fragments := make([]datastore.ByteString, 0)
	chunk := make([]byte, 1024)
	for {
		n, _ := buf.Read(chunk)
		fragments = append(fragments, chunk[:n])
		if n <= 0 {
			break
		}
	}
	ent := &operation{fragments, time.Now()}
	name := ""

	if uniq, ok := x.(Unique); ok {
		name = uniq.Key()
	}

	a := datastore.NewKey(ctx, m.Entity, "root", 0, nil)
	key := datastore.NewKey(ctx, m.Entity, name, 0, a)
	_, err := datastore.Put(ctx, key, ent)
	if err != nil {
		panic(err)
	}
	return err
}
