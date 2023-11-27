package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}

func connectDb() *sql.DB {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
			os.Getenv("PSQL_DB_USER"),
			os.Getenv("PSQL_DB_PWD"),
			os.Getenv("PSQL_DB_NAME"),
			os.Getenv("PSQL_DB_HOST"),
			os.Getenv("PSQL_DB_PORT"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {
	// Connect to the database
	db := connectDb()
	defer db.Close()
	// rail stations list
	railStations(db)
}
