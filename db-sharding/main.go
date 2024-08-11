package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var DBs []*sql.DB

// newConn creates and returns a new database connection to the MySQL server.
// It will panic if the connection cannot be established.
func init() {
	_db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/demo1")
	if err != nil {
		panic(err)
	}
	// Create tables
	_, err = _db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		panic(err)
	}

	_db.Exec("CREATE TABLE users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(255))")

	DBs = append(DBs, _db)

	_db2, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/demo2")
	if err != nil {
		panic(err)
	}

	// Create tables
	_, err = _db2.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		fmt.Println(err)
	}

	_db2.Exec("CREATE TABLE users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(255))")

	DBs = append(DBs, _db2)
	fmt.Println("Database connection established")
}

func shardRouter(UserID int) int {
	var index int = 0
	if UserID%2 == 0 {
		index = 0
	} else {
		index = 1
	}
	return index
}

func main() {
	app := fiber.New()

	// CRUD operations

	// Create a new user
	app.Post("/users", func(c *fiber.Ctx) error {
		user := new(User)
		if err := c.BodyParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Insert user into the database
		_, err := DBs[shardRouter(user.ID)].Exec("INSERT INTO users (id, name) VALUES (?, ?)", user.ID, user.Name)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(user)
	})

	// Get all users
	app.Get("/users", func(c *fiber.Ctx) error {
		users := make([]User, 0)

		// Query all users from the all databases

		for i := 0; i < len(DBs); i++ {
			rows, err := DBs[i].Query("SELECT id, name FROM users")
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			defer rows.Close()

			for rows.Next() {
				user := new(User)
				if err := rows.Scan(&user.ID, &user.Name); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": err.Error(),
					})
				}
				users = append(users, *user)
			}
		}

		return c.JSON(users)
	})

	app.Listen(":3000")
}
