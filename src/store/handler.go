package db

import (
	"encoding/json"
	"log"
)

var (
	Update int = 1
	Create int = 2
	Delete int = 3
)

type Data struct {
	FileID  string
	Content []byte
	Pool    string
}

type Action struct {
	Channel  string
	SenderID string
	Data     Data
	Type     int
}

// Serialize serializes the Action struct to a JSON string.
func (a *Action) Serialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

// Deserialize deserializes a JSON string to an Action struct.
func (a *Action) Deserialize(jsonDaat []byte) error {
	err := json.Unmarshal(jsonDaat, a)
	if err != nil {
		return err
	}
	return nil
}

func (ps *WorkerRoom) handeler(msg Action) {
	if msg.Type == Update {
		ps.update(msg.Data)
	} else if msg.Type == Delete {
		ps.delete(msg.Data)
	} else if msg.Type == Create {
		ps.create(msg.Data)
	}

	//DEBUG
	//log.Println("NEW UPDATE ID:", msg.SenderID[:8], "TYPE:", msg.Type, "ID:", msg.Data.FileID)
}

func (ps *WorkerRoom) update(msg Data) {
	pool, err := ps.db.GetPool(msg.Pool, true)
	if err != nil {
		log.Println("Fail to get pool", err)
		return
	}

	err = pool.Update(msg.FileID, msg.Content)
	if err != nil {
		log.Println("Fail to update record, error:", err)
	}
	event := new(Event)
	event.Name = DbUpdateEvent
	pool.Database.EventBus.Publish(*event)

}

func (ps *WorkerRoom) delete(msg Data) {
	pool, err := ps.db.GetPool(msg.Pool, true)
	if err != nil {
		log.Println("Fail to get pool", err)
		return
	}

	err = pool.Delete(msg.FileID)
	if err != nil {
		log.Println("Fail to delete record, error:", err)
	}
	event := new(Event)
	event.Name = DbUpdateEvent
	pool.Database.EventBus.Publish(*event)

}

func (ps *WorkerRoom) create(msg Data) {
	//get pool
	pool, err := ps.db.GetPool(msg.Pool, true)
	if err != nil {
		log.Println("Fail to get pool", err)
		return
	}

	err = pool.RecordWithID(msg.Content, msg.FileID)
	if err != nil {
		log.Println("Fail to create file, error:", err)
		return
	}
	event := new(Event)
	event.Name = DbUpdateEvent
	pool.Database.EventBus.Publish(*event)
}
