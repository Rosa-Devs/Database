package db

import (
	"context"
	"os"

	"github.com/Rosa-Devs/POC/src/manifest"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DB struct {
	DatabasePath string
}

func (db *DB) Start(path string) {
	db.DatabasePath = path
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
	ctx context.Context

	pb *pubsub.PubSub

	peerId peer.ID

	manifest manifest.Manifest

	db      *DB
	db_name string

	TaskPool chan Action
}

func (db *DB) GetDb(id string, pb *pubsub.PubSub, m manifest.Manifest, peer peer.ID) Database {

	return Database{
		db:       db,
		db_name:  id,
		TaskPool: make(chan Action),
		ctx:      context.Background(),
		pb:       pb, manifest: m,
		peerId: peer,
	}
}

func (db *Database) StartWorker() {
	StartWorker(db)
}

func (db *Database) CreatePool(pool_id string) error {
	pool_path := db.db.DatabasePath + "/" + db.db_name + "/" + pool_id

	err := os.Mkdir(pool_path, 0775)
	if err != nil {
		return err
	}

	return nil
}
