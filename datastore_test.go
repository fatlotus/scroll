package scroll_test

// import (
// 	"github.com/uchicago-sg/scroll"
// 	"google.golang.org/appengine/aetest"
// )

// func ExampleDatastoreLog() {
// 	context, done, err := aetest.NewContext()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer done()
//
// 	log := scroll.DatastoreLog(context, "Entity")
// 	cursor := log.Cursor()
//
// 	read(cursor)
// 	log.Append(Once("peach"))
// 	read(cursor)
//
// 	log.Append(Once("banana"))
// 	log.Append(Once("pear"))
// 	log.Append(Once("banana"))
// 	read(cursor)
// 	read(cursor)
// 	read(cursor)
//
// 	// Output:
// 	// err: scroll.Done: no more entries
// 	// read: peach
// 	// read: pear
// 	// read: banana
// 	// err: scroll.Done: no more entries
// }
