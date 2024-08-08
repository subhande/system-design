package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// newConn creates and returns a new database connection to the MySQL server.
// It will panic if the connection cannot be established.
func newConn() *sql.DB {
	_db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/demo")
	if err != nil {
		panic(err)
	}
	return _db
}

// conn wraps an SQL database connection.
type conn struct {
	db *sql.DB
}

// cpool represents a connection pool with a fixed number of connections.
// The pool manages access to the connections using a mutex and a channel.
type cpool struct {
	mu      *sync.Mutex      // mutex to protect access to the connections slice
	channel chan interface{} // channel to limit the number of concurrent connections
	conns   []*conn          // slice holding the active connections
	maxConn int              // maximum number of connections in the pool
}

// NewCPool creates a new connection pool with a specified maximum number of connections.
// It initializes the pool with connections and returns the pool instance.
func NewCPool(maxConn int) (*cpool, error) {
	var mu = sync.Mutex{}
	pool := &cpool{
		mu:      &mu,
		conns:   make([]*conn, 0, maxConn),
		maxConn: maxConn,
		channel: make(chan interface{}, maxConn),
	}
	for i := 0; i < maxConn; i++ {
		pool.conns = append(pool.conns, &conn{newConn()})
		pool.channel <- nil
	}
	return pool, nil
}

// Close shuts down the connection pool by closing all the connections
// and releasing resources.
func (pool *cpool) Close() {
	close(pool.channel)
	for i := range pool.conns {
		pool.conns[i].db.Close()
	}
}

// Get retrieves a connection from the pool. It blocks until a connection
// becomes available. The connection is removed from the pool until it is returned.
func (pool *cpool) Get() (*conn, error) {
	<-pool.channel
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.conns) == 0 {
		return nil, nil
	}
	c := pool.conns[0]
	pool.conns = pool.conns[1:]
	return c, nil
}

// Put returns a connection to the pool, making it available for reuse.
func (pool *cpool) Put(c *conn) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.conns = append(pool.conns, c)
	pool.channel <- nil
}

// benchmarkPool runs a benchmark to test the performance of the connection pool.
// It spawns n_thread goroutines, each of which gets a connection from the pool,
// executes a dummy query, and returns the connection to the pool.
func benchmarkPool(n_thread int) {
	startTime := time.Now()

	pool, err := NewCPool(10)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < n_thread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := pool.Get()
			if err != nil {
				panic(err)
			}
			_, err = conn.db.Exec("SELECT SLEEP(0.01)") // simulate a query execution with a sleep
			if err != nil {
				panic(err)
			}
			pool.Put(conn)
		}()
	}
	wg.Wait()
	endTime := time.Now()

	fmt.Println("Test: Pool Benchmark | Time Taken: ", endTime.Sub(startTime))
}

// benchmarkPool runs a benchmark to test the performance of the connection pool.
// It spawns n_thread goroutines, each of which gets a connection from the pool,
// executes a dummy query, and returns the connection to the pool.
func benchmarkNonPool(n_thread int) {

	startTime := time.Now()

	var wg sync.WaitGroup

	for i := 0; i < n_thread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db := newConn()
			_, err := db.Query("SELECT SLEEP(0.01)")
			if err != nil {
				panic(err)
			}
			db.Close()
		}()
	}
	wg.Wait()
	endTime := time.Now()
	
	fmt.Println("Test: Non Pool Benchmark | Time Taken: ", endTime.Sub(startTime))
}

func main() {
	//
	n_thread := 10000
	// benchmarkNonPool(n_thread)
	benchmarkPool(n_thread)
	// newConn()
}
