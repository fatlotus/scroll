# Scroll

[![Circle CI](https://img.shields.io/github/issues/fatlotus/scroll.svg)
[![Coveralls](https://img.shields.io/coveralls/fatlotus/scroll.svg
)](https://coveralls.io/github/fatlotus/scroll)

```go
import "github.com/fatlotus/scroll"
```

Scroll is a lightweight log structured database; essentially, it can be
thought of as a distributed implementation []interface{} with fast (and
cheap) scans.

Using it in an application is fairly simple. Generally, you will define your
application as having a series of changes made to a common data structure.
Each of these changes is essentially a database transaction.

(For example, a blog might have BlogEntry and Comment transactions.)

## Unique Entries and Compaction

There's some finesse to using Scroll. Because many transactions are meant
to overwrite earlier entries in the log, you can provide a Key() method that
indicates whether two adjacent entries can be removed.

Suppose we have the following writes to the database.

    A B C D D E A

When reading back the data, Scroll re-orders the events into the following.

    B C D E A

In the blog example above, BlogEntry updates to the same entry would have
the same key, meaning we only see the latest version of an entry.

## Costs on Google App Engine

Scroll is primarily designed for App Engine applications using the App
Engine datastore. The costs for both operations (log.Append and cursor.Next)
are as follows:

    Cost of log.Append: ~(2 + 0.001 * size of payload in bytes) writes.
    Cost of cursor.Next: ~(0.001) reads.

In this case, that payload is encoded using encoding/gob, which anecdotally
performs roughly as an application would on disk.

For an example app with a 10MB working set, that writes 10,000 transactions
of 1KB each, this works out to 10k reads/restore and 30k writes/day.

## Example Application

Suppose we want to represent a basic To-Do list. We start by importing scroll
and x/net/context.

```go
package main

import (
	"fmt"
	"github.com/fatlotus/scroll"
	"golang.org/x/net/context"
)
```

Next, we define the in-memory data model. Nothing here is persisted; instead it
is reconstructed from entries in the log.

```go
// Define the in-memory representation of the application.
type Backend struct {
	Todos  []string
	cursor scroll.Cursor
	log    scroll.Log
}
```

Next, we define _mutations_. These are entries in the log that change the
in-memory state. Importantly, these need to maintain the same schema over time,
since they are stored in a backend database.

```go
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
```

Our application is done! Finally, we add an interface for the user...

```go
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
```

... and start using it from `main()`.

```go
// Add and remove a few items from the to-do list.
func main() {
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
```

## License

Copyright 2016 Jeremy Archer

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.