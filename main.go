package main

import (
	"Sancai/database"
	"Sancai/routers"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var (
	DB  *sql.DB
	err error
)

// local
// const (
// 	host     = "localhost"
// 	port     = "5432"
// 	user     = "postgres"
// 	password = "1234"
// 	dbname   = "sancai_db"
// )

// @title Sancai API
// @version 1.0
// @description API restoran
// @BasePath /api
func main() {
	// local
	// psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
	// 	"password=%s dbname=%s sslmode=disable",
	// 	host, port, user, password, dbname)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"))

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	database.DBMigrate(DB)
	fmt.Println("Successfully connected to database!")
	defer DB.Close()

	router := routers.StartServer(DB)
	router.Run(":" + os.Getenv("PORT"))

	// local
	// router.Run(":8080")
}
