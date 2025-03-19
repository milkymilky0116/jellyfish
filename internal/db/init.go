package db

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitSqliteDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}
