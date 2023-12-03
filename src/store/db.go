package db

import (
	"context"
	"net/http"
	"os"

	"github.com/Rosa-Devs/POC/src/manifest"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DB struct {
	DatabasePath string
	server       *http.Server
	client       *http.Client
	H            host.Host
	Pb           *pubsub.PubSub
	id           peer.ID
}

func (db *DB) Start(path string) {
	db.DatabasePath = path
	db.id = db.H.ID()
	go db.Serve()
}

func (db *DB) CreateDb(m manifest.Manifest) error {

	database_path := db.DatabasePath + "/" + m.Name

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

func (db *DB) GetDb(m manifest.Manifest) Database {

	return Database{
		db:       db,
		db_name:  m.Name,
		TaskPool: make(chan Action),
		ctx:      context.Background(),
		pb:       db.Pb,
		manifest: m,
		peerId:   db.id,
	}
}

func (db *Database) StartWorker() {
	StartWorker(db)
}

func (db *Database) PublishUpdate(a Action) error {
	data := a

	data.SenderID = db.peerId.String()

	db.TaskPool <- data

	//log.Println(data.SenderID)
	return nil

}

func (db *Database) CreatePool(pool_id string) error {
	pool_path := db.db.DatabasePath + "/" + db.db_name + "/" + pool_id

	err := os.Mkdir(pool_path, 0775)
	if err != nil {
		return err
	}

	return nil
}
