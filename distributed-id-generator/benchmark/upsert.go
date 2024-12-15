package benchmark

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/distributed_id_generator/pool"
	"github.com/go-faker/faker/v4"
)

type User3 struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
	Address  string `json:"address"`
}

func PanicPoolClose(err error) {
	// ConnPool2.Close()
	if err != nil {
		panic(err)
	}
}

var ConnPool2 *pool.Cpool

func init() {
	// Get Connection
	connPool, err := pool.NewCPool(50)
	if err != nil {
		PanicPoolClose(err)
	}
	ConnPool2 = connPool
	// fmt.Println("No of connections: ", ConnPool2.GetPoolSize())
	pconn, err := ConnPool2.Get()
	defer ConnPool2.Put(pconn)
	if err != nil {
		PanicPoolClose(err)
	}
	fmt.Println("Connected to the database")

}

func checkIfUserExists(user_id string, pconn *pool.Conn) (bool, User3) {

	rows, err := pconn.Db.Query("SELECT * FROM user2 WHERE id = ?", user_id)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return false, User3{} // Return early and avoid panic
	}
	defer rows.Close() // Ensure rows are closed

	var user User3
	if rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Phone, &user.Website, &user.Address)
		if err != nil {
			fmt.Println("Error scanning rows:", err)
			return false, User3{}
		}
		return true, user
	}
	return false, User3{}
}

func replaceInto(user User3, pconn *pool.Conn) int64 {
	// Parameterized query
	query := `
		REPLACE INTO user2 (id, name, username, email, phone, website, address)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	start := time.Now()

	// Execute the query with parameters
	_, err := pconn.Db.Exec(query, user.ID, user.Name, user.Username, user.Email, user.Phone, user.Website, user.Address)

	duration := time.Since(start).Milliseconds()
	if err != nil {
		fmt.Println("Error executing query:", err)
		PanicPoolClose(err)
	}
	return duration
}

func onDuplicateKeyUpdate(user User3, pconn *pool.Conn) int64 {

	// Parameterized query to prevent SQL injection
	query := `
		INSERT INTO user2 (id, name, username, email, phone, website, address)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE phone = ?`

	start := time.Now()

	// Execute the query with parameters
	_, err := pconn.Db.Exec(query, user.ID, user.Name, user.Username, user.Email, user.Phone, user.Website, user.Address, user.Phone)

	duration := time.Since(start).Milliseconds()
	if err != nil {
		fmt.Println("Error executing query:", err)
	}
	return duration
}

func BenchMarkReplaceIntoVsOnDuplicateKeyUpdate() {
	defer ConnPool2.Close()
	n := 10000

	var wg sync.WaitGroup
	fmt.Println("Benchmarking REPLACE INTO vs ON DUPLICATE KEY UPDATE")
	start := time.Now()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// fmt.Println("No of connections: ", ConnPool2.GetPoolSize())
			pconn, err := ConnPool2.Get()
			defer ConnPool2.Put(pconn)
			if err != nil {
				PanicPoolClose(err)
			}
			// fmt.Println("No of connections: ", ConnPool2.GetPoolSize())

			id, err := rand.Int(rand.Reader, big.NewInt(6000000))
			if err != nil {
				PanicPoolClose(err)
			}
			exists, user := checkIfUserExists(fmt.Sprintf("%d", id), pconn)

			// fmt.Println(exists, user)
			if exists {
				user.Phone = faker.Phonenumber()
				onDuplicateKeyUpdate(user, pconn)
			}
		}()
	}
	wg.Wait()
	end := time.Since(start).Milliseconds()
	// fmt.Println("No of connections: ", ConnPool2.GetPoolSize())
	fmt.Printf("Total time for ON DUPLICATE KEY UPDATE: %d ms\n", end)

	start = time.Now()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// fmt.Println("No of connections: ", ConnPool2.GetPoolSize())
			pconn, err := ConnPool2.Get()
			defer ConnPool2.Put(pconn)
			if err != nil {
				PanicPoolClose(err)
			}
			// fmt.Println("No of connections: ", ConnPool2.GetPoolSize())

			id, err := rand.Int(rand.Reader, big.NewInt(6000000))

			if err != nil {
				PanicPoolClose(err)
			}
			exists, user := checkIfUserExists(fmt.Sprintf("%d", id), pconn)
			// fmt.Println(exists, user)
			if exists {
				user.Phone = faker.Phonenumber()
				replaceInto(user, pconn)

			}
		}()

	}
	wg.Wait()
	end = time.Since(start).Milliseconds()

	fmt.Printf("Total time for REPLACE INTO: %d ms\n", end)

}
