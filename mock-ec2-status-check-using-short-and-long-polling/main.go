package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func init() {

	_db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/demo")

	if err != nil {
		panic(err)
	}

	_, err = _db.Exec("DROP TABLE IF EXISTS servers")
	if err != nil {
		panic(err)
	}

	// _db.Exec("CREATE TABLE servers (server_id INT PRIMARY KEY AUTO_INCREMENT, status VARCHAR(255))")
	_db.Exec("CREATE TABLE servers (server_id INT PRIMARY KEY, status VARCHAR(255))")

	// Insert some dummy data

	_db.Exec("INSERT INTO servers (server_id, status) VALUES (?, ?);", 1, "NOT_YET_CREATED")
	_db.Exec("INSERT INTO servers (server_id, status) VALUES (?, ?);", 2, "NOT_YET_CREATED")
	_db.Exec("INSERT INTO servers (server_id, status) VALUES (?, ?);", 3, "NOT_YET_CREATED")
	_db.Exec("INSERT INTO servers (server_id, status) VALUES (?, ?);", 4, "NOT_YET_CREATED")
	_db.Exec("INSERT INTO servers (server_id, status) VALUES (?, ?);", 5, "NOT_YET_CREATED")

	DB = _db
}

func createEC2(serverID int) {

	waitTime := 5 * time.Second

	fmt.Println("Creating EC2 instance with ID: ", serverID)

	_, err := DB.Exec("UPDATE servers SET status = 'TO_DO' WHERE server_id = ?;", serverID)

	if err != nil {
		panic(err)
	}

	time.Sleep(waitTime)

	_, err = DB.Exec("UPDATE servers SET status = 'IN_PROGRESS' WHERE server_id = ?;", serverID)

	if err != nil {
		panic(err)
	}

	fmt.Println("EC2 instance with ID: ", serverID, " is now in progress")

	time.Sleep(waitTime)

	_, err = DB.Exec("UPDATE servers SET status = 'DONE' WHERE server_id = ?;", serverID)

	if err != nil {
		panic(err)
	}

	fmt.Println("EC2 instance with ID: ", serverID, " is created successfully")

}

func main() {
	ge := gin.Default()

	ge.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Welcome to EC2 instance creation service",
		})
	})

	ge.POST("/create/:server_id", func(ctx *gin.Context) {
		serverID, err := strconv.Atoi(ctx.Param("server_id"))

		if err != nil {
			ctx.JSON(400, gin.H{
				"message": "Invalid server ID",
			})
			return
		}

		go createEC2(serverID)

		ctx.JSON(200, gin.H{
			"message": "EC2 instance creation started",
		})
	})

	ge.GET("/servers", func(ctx *gin.Context) {

		rows, err := DB.Query("SELECT * FROM servers;")

		if err != nil {
			panic(err)
		}

		var servers []gin.H

		for rows.Next() {
			var serverID int
			var status string

			err = rows.Scan(&serverID, &status)

			if err != nil {
				panic(err)
			}

			servers = append(servers, gin.H{
				"server_id": serverID,
				"status":    status,
			})
		}

		ctx.JSON(200, servers)

	})

	ge.GET("/short/status/:server_id", func(ctx *gin.Context) {
		serverID := ctx.Param("server_id")

		var status string

		err := DB.QueryRow("SELECT status FROM servers WHERE server_id = ?;", serverID).Scan(&status)

		if err != nil {
			ctx.JSON(400, gin.H{
				"message": "Invalid server ID",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"server_id": serverID,
			"status":    status,
		})
	})

	ge.GET("/long/status/:server_id", func(ctx *gin.Context) {
		serverID := ctx.Param("server_id")

		rctx := ctx.Request.Context()

		var currentStatus string

		err := DB.QueryRow("SELECT status FROM servers WHERE server_id = ?;", serverID).Scan(&currentStatus)

		if err != nil {
			ctx.JSON(400, gin.H{
				"message": "Invalid server ID",
			})
			return
		}

		var status string

		for {
			// Check if the request is timed out
			if rctx.Err() != nil {
				ctx.JSON(400, gin.H{
					"message": "Request timed out or cancelled",
				})
				return
			}

			err := DB.QueryRow("SELECT status FROM servers WHERE server_id = ?;", serverID).Scan(&status)

			if err != nil {
				ctx.JSON(400, gin.H{
					"message": "Invalid server ID",
				})
				return
			}

			if currentStatus != status {
				break
			}

			time.Sleep(1 * time.Second)

			fmt.Println("Checking status of server: ", serverID, " Current status: ", status)
		}

		ctx.JSON(200, gin.H{
			"server_id": serverID,
			"status":    status,
		})
	})

	ge.Run(":8080")
}
