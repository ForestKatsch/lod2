package db

import (
	"database/sql"
	"errors"
	"log"
)

func ExchangeUserPasswordForToken(username string, password string) (token string, err error) {
	// TODO: implement. for now, return an error
	return "", errors.New("not implemented")
}

// authUsersTable has rows:
// userId TEXT
// userName TEXT
// userPasswordHash TEXT
// userPasswordSalt TEXT
// tid, _ := typeid.WithPrefix("user")

func migrateAuthUsersTable(db *sql.DB, version int) (int, error) {
	if version <= 0 {
		log.Println("migrating auth table from 0 to 1...")
		db.Exec("CREATE TABLE authUsers (userId TEXT PRIMARY KEY, userName TEXT NOT NULL UNIQUE, userPasswordHash TEXT NOT NULL UNIQUE, userPasswordSalt TEXT NOT NULL)")
	}

	return 1, nil
}
