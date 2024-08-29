package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Data struct {
	ID     int    `json:"id"`
	KEY    string `json:"key"`
	VALUE  string `json:"value"`
	EXPIRY int32  `json:"expiry"`
}

func newConn() *sql.DB {
	_db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/store")
	if err != nil {
		panic(err)
	}
	// DROP TABLE
	_, err = _db.Exec("DROP TABLE IF EXISTS kv_store")
	if err != nil {
		panic(err)
	}
	// CREATE TABLE
	_, err = _db.Exec("CREATE TABLE kv_store (id INT AUTO_INCREMENT PRIMARY KEY, k VARCHAR(255) NOT NULL, value TEXT NOT NULL, expiry INT NOT NULL)")
	if err != nil {
		panic(err)

	}
	return _db
}

func set(db *sql.DB, key string, value string, ttl int32) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	expiry := int32(time.Now().Unix()) + ttl
	_, err = tx.Exec("INSERT INTO kv_store (k, value, expiry) VALUES (?, ?, ?)", key, value, expiry)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func get(db *sql.DB, key string) interface{} {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	var data Data
	row := tx.QueryRow("SELECT * FROM kv_store WHERE k = ?", key)
	row.Scan(&data.ID, &data.KEY, &data.VALUE, &data.EXPIRY)
	tx.Commit()
	// check if key exists
	// array
	// d := []interface{}{data.VALUE, data.EXPIRY, data.KEY, data.ID, 45}
	// d := []{data.VALUE, data.EXPIRY, data.KEY, data.ID, 45}
	if data.ID == 0 {
		fmt.Println(key, " not found")
		return nil
	}
	if data.EXPIRY == -1 {
		fmt.Println(key, " has been deleted")
		return nil
	}
	if data.EXPIRY < int32(time.Now().Unix()) {
		fmt.Println(key, " has expired")
		return nil
	}
	// fmt.Println(d)
	fmt.Println(data)
	return data
}

func delete(db *sql.DB, key string) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleting key: ", key)
	// Set expiry to -1 to delete the key
	_, err = tx.Exec("UPDATE kv_store SET expiry = -1 WHERE k = ?", key)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func cleanup(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	timestamp := int32(time.Now().Unix())
	rows, _ := tx.Query("SELECT * FROM kv_store WHERE expiry < ?", timestamp)
	
	for rows.Next() {
		var data Data
		rows.Scan(&data.ID, &data.KEY, &data.VALUE, &data.EXPIRY)
		fmt.Println("Key: ", data.KEY, " will be deleted")
	
	}
	_, err = tx.Exec("DELETE FROM kv_store WHERE expiry < ?", time.Now().Unix())
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func main() {
	db := newConn()
	defer db.Close()
	set(db, "key1", "value1", 30)
	set(db, "key2", "value2", 1)
	set(db, "key3", "value3", 30)
	get(db, "key1")
	get(db, "key2")
	// delete(db, "key1")
	
	// sleep 3 sec
	// time.Sleep(1 * time.Second)
	delete(db, "key3")
	cleanup(db)
	set(db, "key5", "value5", 0)
	time.Sleep(1 * time.Second)
	get(db, "key1")
	get(db, "key2")
	get(db, "key3")
	get(db, "key4")
	get(db, "key5")

}
