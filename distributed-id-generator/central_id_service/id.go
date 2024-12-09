package central_id_service

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// CounterModel represents the counter table in the database
type CounterModel struct {
	ID          int    `json:"id"`
	ServiceName string `json:"service_name"`
	Counter     int64  `json:"counter"`
}

type CounterModel2 struct {
	ID   int    `json:"id"`
	Stub string `json:"stub"`
}

// DBs is a slice of database connections
var DB *sql.DB

// init creates and returns a new database connection to the MySQL server.

func init() {
	_db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/counter_service_db")

	if err != nil {
		panic(err)
	}

	// Create table if not exists
	// service_name is also a unique key
	_, err = _db.Exec("CREATE TABLE IF NOT EXISTS counter (id BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT, service_name VARCHAR(255) UNIQUE, counter BIGINT UNSIGNED NOT NULL)")

	if err != nil {
		panic(err)
	}
	_, err = _db.Exec("CREATE TABLE IF NOT EXISTS tickets (id BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT, stub VARCHAR(255) UNIQUE)")

	if err != nil {
		panic(err)
	}

	DB = _db
	fmt.Println("Database connection established")
}

func GenerateIDAmazon(serviceName string, offset int) []int64 {
	tx, err := DB.Begin()
	if err != nil {
		panic(err)
	}

	var counter CounterModel
	row := tx.QueryRow("SELECT * FROM counter WHERE service_name = ?", serviceName)
	row.Scan(&counter.ID, &counter.ServiceName, &counter.Counter)

	// If the counter does not exist, create a new one
	if counter.ID == 0 {
		_, err = tx.Exec("INSERT INTO counter (service_name, counter) VALUES (?, ?)", serviceName, 0)
		if err != nil {
			panic(err)
		}

		// Retry the query
		row = tx.QueryRow("SELECT * FROM counter WHERE service_name = ?", serviceName)
		row.Scan(&counter.ID, &counter.ServiceName, &counter.Counter)
	}

	// Copy counter value then add 1 to it
	start := counter.Counter + 1

	// Increment the counter by the offset
	counter.Counter += int64(offset)
	_, err = tx.Exec("UPDATE counter SET counter = ? WHERE service_name = ?", counter.Counter, serviceName)
	if err != nil {
		panic(err)
	}

	tx.Commit()

	generatedIDs := []int64{}

	for i := 0; i < offset; i++ {
		generatedIDs = append(generatedIDs, start+int64(i))
	}

	return generatedIDs
}

func GenerateIDFlicker(mode string) int64 {

	tx, err := DB.Begin()
	if err != nil {
		panic(err)
	}

	// var counter CounterModel2

	var generatedID int64

	if mode == "1" {
		result, err := tx.Exec("INSERT INTO tickets (stub) VALUES ('a') ON DUPLICATE KEY UPDATE id = id + 1")
		if err != nil {
			panic(err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}

		generatedID = int64(id)
		
	} else if mode == "2" {
		result, err := tx.Exec("REPLACE INTO tickets (stub) VALUES ('a')")

		if err != nil {
			panic(err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}

		generatedID = int64(id)
	} else {
		_, err = tx.Exec("DELETE FROM tickets WHERE stub = 'a'")

		if err != nil {
			panic(err)
		}

		result, err := tx.Exec("INSERT INTO tickets (stub) VALUES ('a')")
		if err != nil {
			panic(err)
		}

		id, err := result.LastInsertId()

		if err != nil {
			panic(err)
		}

		generatedID = int64(id)
	}

	// row := tx.QueryRow("SELECT * FROM tickets WHERE stub = ?", "a")
	// row.Scan(&counter.ID, &counter.Stub)

	tx.Commit()

	// fmt.Printf("Generated ID: %d, Mode: %s\n", generatedID, mode)

	return int64(generatedID)

}
