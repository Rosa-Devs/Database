package db

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Rosa-Devs/Database/src/chiper"
	"github.com/Rosa-Devs/Database/src/manifest"
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

	database_path := db.DatabasePath + "/" + m.UId
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

	db *DB

	TaskPool chan Action

	chiper *chiper.Chiper

	EventBus EventBus
}

func (db *DB) GetDb(m manifest.Manifest) Database {

	c, err := chiper.NewChiper(m.Chiper)
	if err != nil {
		log.Println("Failed to create chiper!!!")
		return Database{}
	}

	return Database{
		db:       db,
		TaskPool: make(chan Action),
		ctx:      context.Background(),
		pb:       db.Pb,
		manifest: m,
		peerId:   db.id,
		EventBus: *NewEventBus(),
		chiper:   c,
	}
}

func (db *Database) StartWorker(timeout int) {
	StartWorker(db, timeout)
}

func (db *Database) PublishUpdate(a Action) error {
	data := a

	data.SenderID = db.peerId.String()

	db.TaskPool <- data

	//log.Println(data.SenderID)
	return nil

}

func (db *Database) CreatePool(pool_id string) error {
	pool_path := db.db.DatabasePath + "/" + db.manifest.UId + "/" + pool_id

	err := os.Mkdir(pool_path, 0775)
	if err != nil {
		return err
	}

	return nil
}
