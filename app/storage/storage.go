package storage

import (
	"time"
)

type Db struct {
	hash map[string]any
	list []any
}

func NewDb() Db {
	db := Db{hash: map[string]any{}, list: []any{}}
	return db
}

func (db *Db) Set(k string, v any, px int64) {
	db.hash[k] = v
	if px != -1 {
		go clearAfterInterval(px, k, db)
	}
}

func (db *Db) Get(k string) any {
	return db.hash[k]
}

func clearAfterInterval(px int64, k string, db *Db) {
	time.Sleep(time.Duration(px) * time.Millisecond)
	delete(db.hash, k)
}
