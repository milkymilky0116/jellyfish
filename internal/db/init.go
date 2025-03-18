package db

import (
	"github.com/dgraph-io/badger/v4"
)

func InitBadgerDB() (*badger.DB, error) {
	db, err := badger.Open(badger.DefaultOptions("."))
	if err != nil {
		return nil, err
	}
	return db, nil
}
