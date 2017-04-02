package main

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	db, _ := sql.Open("postgres", "user=seedbox dbname=seedbox host=direct.xserv.co.za password=dezrcq500 port=5433 sslmode=disable")
	db.Ping()
}
