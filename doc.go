// Scroll is a lightweight log structured database; essentially, it can be
// thought of as a distributed implementation []interface{} with fast (and
// cheap) scans.
//
// Using it in an application is fairly simple. Generally, you will define your
// application as having a series of changes made to a common data structure.
// Each of these changes is essentially a database transaction.
//
// (For example, a blog might have BlogEntry and Comment transactions.)
//
// Unique Entries and Compaction
//
// There's some finesse to using Scroll. Because many transactions are meant
// to overwrite earlier entries in the log, you can provide a Key() method that
// indicates whether two adjacent entries can be removed.
//
// Suppose we have the following writes to the database.
//
//  A B C D D E A
//
// When reading back the data, Scroll re-orders the events into the following.
//
//  B C D E A
//
// In the blog example above, BlogEntry updates to the same entry would have
// the same key, meaning we only see the latest version of an entry.
//
// Costs on Google App Engine
//
// Scroll is primarily designed for App Engine applications using the App
// Engine datastore. The costs for both operations (log.Append and cursor.Next)
// are as follows:
//
//  Cost of log.Append: ~(2 + 0.001 * size of payload in bytes) writes.
//  Cost of cursor.Next: ~(0.001) reads.
//
// In this case, that payload is encoded using encoding/gob, which anecdotally
// performs roughly as an application would on disk.
//
// For an example app with a 10MB working set, that writes 10,000 transactions
// of 1KB each, this works out to 10k reads/restore and 30k writes/day.
package scroll
