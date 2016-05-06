package scroll

import (
	"encoding/json"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"strings"
	"sync"
	"time"
)

type operation struct {
	Data  []string
	Order time.Time
}

type dbLog struct {
	Entity  string
	Query   *datastore.Query
	Context context.Context
	sync.Mutex
}

type dbCursor struct {
	LastTime time.Time
	Log      *dbLog
	Pending  []operation
}

// A DatastoreLog is a Log stored on the Google App Engine Datastore, a
// wrapper around BigTable.
//
//  Cost of log.Append: ~(2 + 0.001 * size of gob payload in bytes) writes.
//  Cost of cursor.Next: ~(0.001) reads.
func DatastoreLog(entity string) Log {
	q := datastore.NewQuery(entity).Order("Order")
	return &dbLog{
		Entity: entity,
		Query:  q,
	}
}

func (m *dbLog) Cursor() Cursor {
	return &dbCursor{
		Log: m,
	}
}

func (c *dbCursor) Next(ctx context.Context, x interface{}) error {
	c.Log.Lock()
	defer c.Log.Unlock()

	if len(c.Pending) == 0 {
		c.Pending = make([]operation, 0)
		q := c.Log.Query.Filter("Order >", c.LastTime).Limit(1000)
		_, err := q.GetAll(ctx, &c.Pending)
		if err != nil {
			return err
		} else if len(c.Pending) == 0 {
			return Done
		}
	}

	op := c.Pending[0]
	c.Pending, c.LastTime = c.Pending[1:], op.Order
	return json.Unmarshal([]byte(strings.Join(op.Data, "")), x)
}

func (m *dbLog) Append(ctx context.Context, x interface{}) error {
	m.Lock()
	defer m.Unlock()

	data, err := json.Marshal(x)
	if err != nil {
		return err
	}
	fragments := make([]string, 0)
	for i := 0; i < len(data); i += 1024 {
		fragments = append(fragments, string(data[i:i+1024]))
	}
	ent := &operation{fragments, time.Now()}
	name := ""

	if uniq, ok := x.(Unique); ok {
		name = uniq.Key()
	}
	key := datastore.NewKey(ctx, m.Entity, name, 0, nil)
	_, err = datastore.Put(ctx, key, ent)
	return err
}
