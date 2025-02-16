package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "lod2.db")

	if err != nil {
		log.Fatal("unable to open db", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal("unable to ping db", err)
	}
}
