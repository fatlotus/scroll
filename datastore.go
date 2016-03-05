package scroll

import (
	"encoding/json"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"sync"
	"time"
)

type Operation struct {
	Data  []byte
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
	Pending  []Operation
}

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
		c.Pending = make([]Operation, 0)
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
	return json.Unmarshal(op.Data, x)
}

func (m *dbLog) Append(ctx context.Context, x interface{}) error {
	m.Lock()
	defer m.Unlock()

	data, err := json.Marshal(x)
	if err != nil {
		return err
	}
	ent := &Operation{data, time.Now()}
	name := ""

	if uniq, ok := x.(Unique); ok {
		name = uniq.Key()
	}
	key := datastore.NewKey(ctx, m.Entity, name, 0, nil)
	_, err = datastore.Put(ctx, key, ent)
	return err
}
