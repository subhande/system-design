package pool

import (
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// newConn creates and returns a new database connection to the MySQL server.
// It will panic if the connection cannot be established.
func newConn() *sql.DB {
	_db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/counter_service_db")
	if err != nil {
		panic(err)
	}
	return _db
}

// Conn wraps an SQL database connection.
type Conn struct {
	Db *sql.DB
}

// Cpool represents a connection pool with a fixed number of connections.
// The pool manages access to the connections using a mutex and a channel.
type Cpool struct {
	mu      *sync.Mutex      // mutex to protect access to the connections slice
	channel chan interface{} // channel to limit the number of concurrent connections
	conns   []*Conn          // slice holding the active connections
	maxConn int              // maximum number of connections in the pool
}

// NewCPool creates a new connection pool with a specified maximum number of connections.
// It initializes the pool with connections and returns the pool instance.
func NewCPool(maxConn int) (*Cpool, error) {
	var mu = sync.Mutex{}
	pool := &Cpool{
		mu:      &mu,
		conns:   make([]*Conn, 0, maxConn),
		maxConn: maxConn,
		channel: make(chan interface{}, maxConn),
	}
	for i := 0; i < maxConn; i++ {
		pool.conns = append(pool.conns, &Conn{newConn()})
		pool.channel <- nil
	}
	return pool, nil
}

// Close shuts down the connection pool by closing all the connections
// and releasing resources.
func (pool *Cpool) Close() {
	close(pool.channel)
	for i := range pool.conns {
		pool.conns[i].Db.Close()
	}

}

// Get retrieves a connection from the pool. It blocks until a connection
// becomes available. The connection is removed from the pool until it is returned.
func (pool *Cpool) Get() (*Conn, error) {
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

func (p *Cpool) IsChannelClosed() bool {
    select {
    case _, ok := <-p.channel:
        return !ok // If `ok` is false, the channel is closed.
    default:
        return false // If no value is available, assume the channel is open.
    }
}

func (p *Cpool) GetPoolSize() int {
	return len(p.conns)
}

// Put returns a connection to the pool, making it available for reuse.
func (pool *Cpool) Put(c *Conn) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.conns = append(pool.conns, c)
	if !pool.IsChannelClosed() {
		pool.channel <- nil
	}
}
