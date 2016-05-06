package scroll

import (
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
	"testing"
)

type Entity string

func (r Entity) Key() string { return string(r) }

func RunCursor(t *testing.T, c context.Context, cur Cursor, e []string) {
	var record Entity
	for i := 0; i < len(e); i++ {
		err := cur.Next(c, &record)
		if err != nil {
			t.Errorf("got unexpected error: %s\n", err)
		} else if string(record) != e[i] {
			t.Errorf("expecting %s, got %s\n", e[i], record)
		}
	}
	if err := cur.Next(c, &record); err != Done {
		t.Errorf("got (%s, err=%s), expecting end of list.\n", record, err)
	}
}

func RunLog(t *testing.T, ctx context.Context, l Log) {
	// Test normal insertion.
	c := l.Cursor()
	RunCursor(t, ctx, c, []string{})
	l.Append(ctx, Entity("strawberry"))
	l.Append(ctx, Entity("banana"))
	RunCursor(t, ctx, c, []string{"strawberry", "banana"})

	// Test duplicated unique keys.
	l.Append(ctx, Entity("pear"))
	l.Append(ctx, Entity("pear"))
	l.Append(ctx, Entity("grape"))
	RunCursor(t, ctx, c, []string{"pear", "grape"})

	// Test re-ordering keys.
	l.Append(ctx, Entity("pear"))
	RunCursor(t, ctx, c, []string{"pear"})

	c = l.Cursor()
	RunCursor(t, ctx, c, []string{"strawberry", "banana", "grape", "pear"})
}

func TestMemoryLog(t *testing.T) {
	ctx := context.Background()
	RunLog(t, ctx, MemoryLog())
}

func TestFileLog(t *testing.T) {
	ctx := context.Background()
	fp, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(fp.Name())
	defer fp.Close()
	RunLog(t, ctx, FileLog(fp.Name()))
}
