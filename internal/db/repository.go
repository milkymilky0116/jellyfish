package db

import (
	"slices"

	"github.com/dgraph-io/badger/v4"
)

type IRepository interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
}

type BadgerRepository struct {
	DB *badger.DB
}

func (b *BadgerRepository) Get(key string) ([]byte, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			value = slices.Clone(val)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (b *BadgerRepository) Set(key string, value []byte) error {
	err := b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		return err
	}
	return nil
}

func InitBadgerRepository(db *badger.DB) IRepository {
	return &BadgerRepository{DB: db}
}
