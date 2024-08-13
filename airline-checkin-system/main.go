package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-faker/faker/v4"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Trip struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Seat struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	UserID int    `json:"user_id"`
	TripID int    `json:"trip_id"`
}

func newConn() *sql.DB {
	_db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/airline")
	if err != nil {
		panic(err)
	}
	return _db
}

func insert_data(n int) {
	db := newConn()
	defer db.Close()
	// Create tables
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Exec("DROP TABLE IF EXISTS trips")
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Exec("DROP TABLE IF EXISTS seats")
	if err != nil {
		fmt.Println(err)
	}
	db.Exec("CREATE TABLE users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(255))")
	db.Exec("CREATE TABLE trips (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(255))")
	db.Exec("CREATE TABLE seats (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(255), user_id INT, trip_id INT)")
	for i := 0; i < n; i++ {
		user := User{
			ID:   i + 1,
			Name: faker.Name(),
		}
		_, err := db.Exec("INSERT INTO users (name) VALUES (?)", user.Name)
		if err != nil {
			panic(err)
		}
	}

	// Insert trips
	_, err = db.Exec("INSERT INTO trips (name) VALUES (?)", "AIRINDIA-101")
	if err != nil {
		panic(err)
	}

	// Insert seats
	// Airplane Seats
	// Convert row to int row / 6
	map_seat := map[int]string{
		0: "A",
		1: "B",
		2: "C",
		3: "D",
		4: "E",
		5: "F",
	}
	row := int(n / 6)
	for i := 0; i < row; i++ {
		for j := 0; j < 6; j++ {
			seat := strconv.Itoa(i+1) + "-" + map_seat[j]
			_, err := db.Exec("INSERT INTO seats (name, user_id, trip_id) VALUES (?, ?, ?)", seat, nil, 1)
			if err != nil {
				panic(err)
			}
		}

	}

}

func assign_seat_sequesntially(n int) {
	start := time.Now()
	db := newConn()
	defer db.Close()
	for i := 0; i < n; i++ {
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		// Get 1st available seat
		row := tx.QueryRow("SELECT id, name, user_id, trip_id FROM seats WHERE trip_id = 1 AND user_id IS NULL LIMIT 1")

		var seat Seat
		row.Scan(&seat.ID, &seat.Name, &seat.UserID, &seat.TripID)
		// fmt.Println(seat)

		// Assign seat to user
		_, err = tx.Exec("UPDATE seats SET user_id = ? WHERE id = ?", i+1, seat.ID)
		if err != nil {
			panic(err)
		}
		tx.Commit()

	}
	elapsed := time.Since(start)
	fmt.Println("Time taken to assign seats sequentially: ", elapsed)
}

func assign_seat_without_lock(n int) {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(user_id int, wg *sync.WaitGroup) {
			db := newConn()
			// defer db.Close()
			defer wg.Done()
			// Start transaction
			tx, err := db.Begin()
			if err != nil {
				// panic(err)
				fmt.Println(err, user_id)
				db.Close()
			}
			row := tx.QueryRow("SELECT id, name, user_id, trip_id FROM seats WHERE trip_id = 1 AND user_id IS NULL LIMIT 1")

			var seat Seat
			row.Scan(&seat.ID, &seat.Name, &seat.UserID, &seat.TripID)
			// fmt.Println(seat)

			// Assign seat to user
			_, err = tx.Exec("UPDATE seats SET user_id = ? WHERE id = ?", user_id, seat.ID)
			if err != nil {
				db.Close()
				// panic(err)
				fmt.Println(err, user_id)
			}
			tx.Commit()
			db.Close()
		}(i+1, &wg)

	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Time taken to assign seats without lock: ", elapsed)
}

func assign_seat_with_lock(n int) {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(user_id int, wg *sync.WaitGroup) {
			db := newConn()
			// defer db.Close()
			defer wg.Done()
			// Start transaction
			tx, err := db.Begin()
			if err != nil {
				// panic(err)
				fmt.Println(err, user_id)
				db.Close()
			}
			// Take Exclusive lock
			row := tx.QueryRow("SELECT id, name, user_id, trip_id FROM seats WHERE trip_id = 1 AND user_id IS NULL LIMIT 1 FOR UPDATE")

			var seat Seat
			row.Scan(&seat.ID, &seat.Name, &seat.UserID, &seat.TripID)
			// fmt.Println(seat)

			// Assign seat to user
			_, err = tx.Exec("UPDATE seats SET user_id = ? WHERE id = ?", user_id, seat.ID)
			if err != nil {
				db.Close()
				// panic(err)
				fmt.Println(err, user_id)
			}
			tx.Commit()
			db.Close()
		}(i+1, &wg)

	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Time taken to assign seats with lock: ", elapsed)
}

func assign_seat_with_skip_lock(n int) {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(user_id int, wg *sync.WaitGroup) {
			db := newConn()
			// defer db.Close()
			defer wg.Done()
			// Start transaction
			tx, err := db.Begin()
			if err != nil {
				// panic(err)
				fmt.Println(err, user_id)
				db.Close()
			}
			// Take Exclusive lock
			row := tx.QueryRow("SELECT id, name, user_id, trip_id FROM seats WHERE trip_id = 1 AND user_id IS NULL LIMIT 1 FOR UPDATE SKIP LOCKED")

			var seat Seat
			row.Scan(&seat.ID, &seat.Name, &seat.UserID, &seat.TripID)
			// fmt.Println(seat)

			// Assign seat to user
			_, err = tx.Exec("UPDATE seats SET user_id = ? WHERE id = ?", user_id, seat.ID)
			if err != nil {
				db.Close()
				// panic(err)
				fmt.Println(err, user_id)
			}
			tx.Commit()
			db.Close()
		}(i+1, &wg)

	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Time taken to assign seats with skip lock: ", elapsed)
}

func printSeats(n int) {
	
	db := newConn()
	defer db.Close()
	rows, err := db.Query("SELECT id, name, user_id, trip_id FROM seats WHERE trip_id = 1 AND user_id IS NOT NULL")
	if err != nil {
		panic(err)
	}
	count := 0
	defer rows.Close()
	for rows.Next() {
		var seat Seat
		rows.Scan(&seat.ID, &seat.Name, &seat.UserID, &seat.TripID)
		// fmt.Println(seat)
		count++
	}
	fmt.Println("Total seats assigned: ", count)
	fmt.Println()
	for i := 0; i < n; i++ {
		if i%int(n/6) == 0 {
			fmt.Println()
		}
		if i <= count {
			fmt.Print(" * ")
		} else {
			fmt.Print(" x ")
		}

	}
	fmt.Println()
	fmt.Println()
	fmt.Println("==================================================")
}

func main() {
	var n int = 120
	insert_data(n)
	assign_seat_sequesntially(n)
	printSeats(n)

	insert_data(n)
	assign_seat_without_lock(n)
	printSeats(n)

	insert_data(n)
	assign_seat_with_lock(n)
	printSeats(n)

	insert_data(n)
	assign_seat_with_skip_lock(n)
	printSeats(n)
}
