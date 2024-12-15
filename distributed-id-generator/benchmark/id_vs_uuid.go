package benchmark

import (
	"fmt"
	"sync"
	"time"

	"github.com/distributed_id_generator/pool"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

var User1 struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
	Address  string `json:"address"`
}

var User2 struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
	Address  string `json:"address"`
}

var ConnPool *pool.Cpool

func init() {
	// Get Connection
	connPool, err := pool.NewCPool(1)
	if err != nil {
		panic(err)
	}
	ConnPool = connPool

	conn, err := ConnPool.Get()
	if err != nil {
		panic(err)
	}
	defer ConnPool.Put(conn)

	// // Drop table user 1 and user 2 if exists

	// query_del := `
	// 	DROP TABLE IF EXISTS user1
	// `
	// _, err = conn.Db.Exec(query_del)
	// if err != nil {
	// 	panic(err)
	// }

	// query_del = `
	// 	DROP TABLE IF EXISTS user2
	// `
	// _, err = conn.Db.Exec(query_del)
	// if err != nil {
	// 	panic(err)
	// }

	// Create table user 1 and user 2 if not exists
	query := `
		CREATE TABLE IF NOT EXISTS user1 (
		 	id VARCHAR(255) NOT NULL PRIMARY KEY,
			name VARCHAR(255),
			username VARCHAR(255),
			email VARCHAR(255),
			phone VARCHAR(255),
			website VARCHAR(255),
			address VARCHAR(255)
		)
	`
	_, err = conn.Db.Exec(query)
	if err != nil {
		panic(err)
	}

	query = `
		CREATE TABLE IF NOT EXISTS user2 (
		 id BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
		 name VARCHAR(255),
		 username VARCHAR(255),
		 email VARCHAR(255),
		 phone VARCHAR(255),
		 website VARCHAR(255),
		 address VARCHAR(255)
		)
	`

	_, err = conn.Db.Exec(query)

	if err != nil {
		panic(err)
	}
}

func insertUser1() int64 {
	conn, err := ConnPool.Get()
	if err != nil {
		panic(err)
	}
	defer ConnPool.Put(conn)

	// Go generate UUID
	generatedUUID := uuid.New().String()
	id := generatedUUID
	name := faker.Name()
	username := faker.Username()
	email := faker.Email()
	phone := faker.Phonenumber()
	website := faker.URL()
	realAddress := faker.GetRealAddress()

	address := realAddress.Address + ", " + realAddress.City + ", " + realAddress.State + ", " + realAddress.PostalCode

	start := time.Now()
	query := `
		INSERT INTO user1 (id, name, username, email, phone, website, address)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = conn.Db.Exec(query, id, name, username, email, phone, website, address)

	duration := time.Since(start).Milliseconds()

	if err != nil {
		panic(err)
	}

	return duration

}

func insertUser2() int64 {
	conn, err := ConnPool.Get()
	if err != nil {
		panic(err)
	}
	defer ConnPool.Put(conn)

	name := faker.Name()
	username := faker.Username()
	email := faker.Email()
	phone := faker.Phonenumber()
	website := faker.URL()
	realAddress := faker.GetRealAddress()

	address := realAddress.Address + ", " + realAddress.City + ", " + realAddress.State + ", " + realAddress.PostalCode

	start := time.Now()
	query := `
		INSERT INTO user2 (name, username, email, phone, website, address)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = conn.Db.Exec(query, name, username, email, phone, website, address)

	if err != nil {
		panic(err)
	}

	duration := time.Since(start).Milliseconds()

	return duration

}

func InsertBulk() {
	NO_OF_USERS := 1000
	// BATCH_SIZE := 1000
	var wg sync.WaitGroup // WaitGroup for goroutines
	var totalDuration int64
	var avgDuration int64 = 0
	totalDuration = 0
	start := time.Now()
	for i := 0; i < NO_OF_USERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// duration := insertUser1()
			// totalDuration += duration
			insertUser1()
		}()
	}
	wg.Wait()
	// avgDuration := totalDuration / int64(NO_OF_USERS)
	totalDuration = time.Since(start).Milliseconds()
	fmt.Printf("User1 | Inserted: %d | Avg Duration: %d ms | Total Duration: %d ms\n", NO_OF_USERS, avgDuration, totalDuration)
	totalDuration = 0
	start = time.Now()
	for i := 0; i < NO_OF_USERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// duration := insertUser2()
			// totalDuration += duration
			insertUser2()
		}()

	}
	wg.Wait()
	// avgDuration = totalDuration / int64(NO_OF_USERS)
	totalDuration = time.Since(start).Milliseconds()
	fmt.Printf("User2 | Inserted: %d | Avg Duration: %d ms | Total Duration: %d ms\n", NO_OF_USERS, avgDuration, totalDuration)
}

func InsertBulkTest() {
	fmt.Println("=========================================")
	for i := 0; i < 6; i++ {
		fmt.Printf("Already inserted %dM records into user1 and user2 | Inserting next 1M records\n", i)
		InsertBulk()
	}
}
