package db

import (
	"errors"
)

func init() {
	//db.Begin()
}

func ExchangeUserPasswordForToken(username string, password string) (token string, err error) {
	// TODO: implement. for now, return an error
	return "", errors.New("not implemented")
}
