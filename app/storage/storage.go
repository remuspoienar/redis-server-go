package storage

type Db struct {
	hash map[string]any
	list []any
}

func NewDb() Db {
	db := Db{hash: map[string]any{}, list: []any{}}
	return db
}

func (db *Db) Set(k string, v any) {
	db.hash[k] = v
}

func (db *Db) Get(k string) any {
	return db.hash[k]
}
