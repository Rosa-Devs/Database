package db

import (
	"os"

	"github.com/Rosa-Devs/POC/src/p2p"
)

type DB struct {
	DatabasePath string
	ChatRoom     p2p.ChatRoom
}

func (db *DB) Start(path string) {

	db.DatabasePath = path
	spawnWorkers()
}

func spawnWorkers() {

}

func (db *DB) CreateDb(id string) error {

	database_path := db.DatabasePath + "/" + id

	err := os.MkdirAll(database_path, 0775)
	if err != nil {
		return err
	}

	return nil
}

type Database struct {
	db      *DB
	db_name string
}

func (db *DB) GetDb(id string) Database {
	return Database{db: db, db_name: id}
}

func (db *Database) CreatePool(pool_id string) error {
	pool_path := db.db.DatabasePath + "/" + db.db_name + "/" + pool_id

	err := os.Mkdir(pool_path, 0775)
	if err != nil {
		return err
	}

	return nil
}
