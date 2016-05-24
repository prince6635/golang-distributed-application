package datamanager

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	// db, err = sql.Open(
	// 	"postgres", //driverName string,
	// 	"postgres://distributed:distributed@localhost/distributed?sslmode=disable") //dataSourceName string)
	db, err = sql.Open("postgres", "user=distributed password=admin dbname=distributed sslmode=disable")

	if err != nil {
		panic(err.Error())
	}
}
