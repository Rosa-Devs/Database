package db

import (
	"context"
	"encoding/json"
	"log"

	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 1572864

// shortID returns the last 8 chars of a base58-encoded peer id.

func ShortID(p peer.ID) string {
	pretty := p.String()
	return pretty[len(pretty)-8:]
}

type WorkerRoom struct {

	//Database
	db *Database
	//Db Manifest

	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *Action

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID

	timeout int
}

func StartWorker(db *Database, timeout int) (*WorkerRoom, error) {
	// join the pubsub topic
	topic, err := db.pb.Join(db.manifest.PubSub)
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &WorkerRoom{
		ctx:      db.ctx,
		ps:       db.pb,
		topic:    topic,
		sub:      sub,
		self:     db.peerId,
		roomName: db.manifest.PubSub,
		Messages: make(chan *Action, ChatRoomBufSize),
		db:       db,
		timeout:  timeout,
	}

	// start reading messages from the subscription in a loop
	log.Println("Starting workers")
	go cr.Sync()
	go cr.TaskPublisher()
	go cr.readLoop()
	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *WorkerRoom) TaskPublisher() {
	for {
		for task := range cr.db.TaskPool {
			if task.Channel == cr.roomName {
				task_data, err := task.Serialize()
				if err != nil {
					log.Println(err)
					continue
				}

				// Encrypt data
				edata, err := cr.db.chiper.Encrypt(task_data)
				if err != nil {
					log.Println("Failed to ecnrypt message!", err)
					continue
				}
				err = cr.topic.Publish(cr.ctx, edata)
				if err != nil {
					log.Println("Error publishing task:", err)
					continue
				}
			}

		}
	}
}

func (cr *WorkerRoom) ListPeers() []peer.ID {
	return cr.ps.ListPeers(cr.roomName)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (cr *WorkerRoom) readLoop() {
	for {
		msg, err := cr.sub.Next(cr.ctx)
		if err != nil {
			continue
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == cr.self {
			continue
		}

		cm := new(Action)

		edata, err := cr.db.chiper.Decrypt(msg.Data)
		if err != nil {
			log.Println("Failed to decrypt message!", err)
		}

		err = json.Unmarshal(edata, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		go cr.handeler(*cm)
	}
}
