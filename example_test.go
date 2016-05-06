package scroll_test

import (
	"fmt"
	"github.com/fatlotus/scroll"
	"golang.org/x/net/context"
)

// Define what a mutation looks like.
type Mutation interface {
	Update(b *Backend)
}

// Define the types of mutations to store in the log.
type AddItem string
type RemoveItem string

// Make sure duplicate operations are merged together.
func (a AddItem) Key() string    { return string(a) }
func (r RemoveItem) Key() string { return string(r) }

func (a AddItem) Update(b *Backend) {
	fmt.Printf("AddItem(%s)\n", a)
	for _, x := range b.Todos {
		if string(a) == x {
			return
		}
	}
	b.Todos = append(b.Todos, string(a))
}

func (r RemoveItem) Update(b *Backend) {
	fmt.Printf("RemoveItem(%s)\n", r)
	j := 0
	for i := 0; i < len(b.Todos); i++ {
		if b.Todos[i] != string(r) {
			b.Todos[j] = b.Todos[i]
			j++
		}
	}
	b.Todos = b.Todos[:j]
}

// Define the in-memory representation of the application.
type Backend struct {
	Todos  []string
	cursor scroll.Cursor
	log    scroll.Log
}

// Pull the latest versions from Scroll.
func (b *Backend) Update() (c context.Context, err error) {
	var m Mutation
	for {
		err = b.cursor.Next(c, &m)
		if err == scroll.Done {
			break
		} else if err != nil {
			return
		}
		m.Update(b)
	}
	return
}

// Define an interface for clients to use.
func (b *Backend) Add(c context.Context, item string) {
	b.log.Append(c, AddItem(item))
}

func (b *Backend) Remove(c context.Context, item string) {
	b.log.Append(c, RemoveItem(item))
}

func NewBackend() *Backend {
	log := scroll.MemoryLog()
	cursor := log.Cursor()
	return &Backend{Todos: make([]string, 0), log: log, cursor: cursor}
}

// Add and remove a few items from the to-do list.
func Example() {
	c := context.Background()

	b := NewBackend()
	b.Add(c, "apples")
	b.Add(c, "bananas")
	b.Add(c, "pears")
	b.Add(c, "bananas")
	b.Remove(c, "pears")

	b.Update()

	fmt.Printf("b.Todos: %s\n", b.Todos)

	// Output:
	// AddItem(apples)
	// AddItem(bananas)
	// RemoveItem(pears)
	// b.Todos: [apples bananas]
}
