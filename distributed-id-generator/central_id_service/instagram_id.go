package central_id_service

import (
	"fmt"
	"log"

	"github.com/go-faker/faker/v4"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5"
)

var DBPG *pgx.Conn

func getConnection(database string) *pgx.Conn {
	if database == "" {
		database = "demo"
	}
	config := pgx.ConnConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "root",
		Database: database,
	}

	// Establish a connection
	_db, err := pgx.Connect(config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return _db
}

// init creates and returns a new database connection to the MySQL server.
func init() {
	// Database connection URL (replace with your credentials)
	// connStr := "postgres://postgres:root@localhost:5432/demo"

	_db := getConnection("")

	DB_PREFIX := "insta"

	// for  1-15
	for i := 1; i <= 15; i++ {
		DB_ID := i
		DB_NAME := DB_PREFIX + fmt.Sprintf("_%d", DB_ID)
		// Create the database
		_, err := _db.Exec("CREATE DATABASE " + DB_NAME)

		if err != nil {
			fmt.Println(err)
		}

		_db = getConnection(DB_NAME)

		_, err = _db.Exec(fmt.Sprintf(`CREATE SCHEMA %s;`, DB_NAME))

		if err != nil {
			fmt.Println(err)
		}

		_, err = _db.Exec(fmt.Sprintf(`CREATE SEQUENCE %s.table_id_seq START 1;`, DB_NAME))

		if err != nil {
			fmt.Println(err)
		}

		// Add stored procedure
		// multi-line string
		storedProc := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION
			%s.next_id(OUT result BIGINT) AS $$
		DECLARE
			epoch BIGINT := 1293840000000; -- 2011-01-01 00:00:00
			seq_id BIGINT;
			now_ms BIGINT;
			shard_id INT := %d;
		BEGIN
			-- Get sequence number
			SELECT nextval('%s.table_id_seq') %% 1024 INTO seq_id;

			-- Get the current timestamp in milliseconds
			now_ms := FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000);

			-- Calculate the result
			result := (now_ms - epoch) << 23;
			result := result | (shard_id << 10);
			result := result | seq_id;
		END;
		$$ LANGUAGE plpgsql;
		`, DB_NAME, DB_ID, DB_NAME)

		_, err = _db.Exec(storedProc)
		if err != nil {
			panic(err)
		}

		// Create table if not exists
		_, err = _db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS posts (id BIGINT NOT NULL DEFAULT %s.next_id(), data VARCHAR(255))", DB_NAME))
		if err != nil {
			panic(err)
		}

		DBPG = _db
	}

	fmt.Println("Database connection established")
}

// GenerateIDInstagram generates a new ID for Instagram.

func GenerateIDSnowFlakeInstagram() []int64 {

	ids := []int64{}

	// genrate random string of 6 characters
	data := faker.Username()

	// Insert the data into the table

	for i := 1; i <= 15; i++ {
		var id int64

		DB_NAME := "insta" + fmt.Sprintf("_%d", i)

		_db := getConnection(DB_NAME)

		for j := 0; j < 10; j++ {
			err := _db.QueryRow("INSERT INTO posts (data) VALUES ($1) RETURNING id", data).Scan(&id)

			if err != nil {
				panic(err)
			}
			ids = append(ids, id)
			// fmt.Println("Inserted ID:", id, " into ", DB_NAME)
		}
		_, err := _db.Query("SELECT * FROM posts")
		if err != nil {
			panic(err)
		}

		// for result.Next() {
		// 	var data string
		// 	var id int64
		// 	err = result.Scan(&id, &data)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	fmt.Println("ID:", id, " Data:", data, " from ", DB_NAME)
		// }

	}

	return ids
}
