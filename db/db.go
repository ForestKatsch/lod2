package db

import (
	"database/sql"
	"lod2/config"
	"lod2/utils"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var DB *sql.DB

func Init() {
	dbPath := filepath.Join(config.Config.DataPath, "lod2.db")
	utils.EnsureDirForFile(dbPath)

	var err error
	db, err = sql.Open("sqlite3", dbPath)

	if err != nil {
		log.Fatal("unable to open db: ", err)
	}

	// Ping the db to make sure it works
	err = db.Ping()

	if err != nil {
		log.Fatal("unable to ping db: ", err)
	}

	log.Printf("db opened at '%s'", dbPath)

	DB = db
}

