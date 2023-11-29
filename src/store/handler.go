package db

import "encoding/json"

var (
	Update int = 1
	Create int = 2
	Delete int = 3
)

type Data struct {
	FileID  string
	Content []byte
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
func (a *Action) Deserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), a)
	if err != nil {
		return err
	}
	return nil
}

func handeler(msg Action) {
	if msg.Type == Update {
		update(msg.Data)
	} else if msg.Type == Delete {
		delete(msg.Data)
	} else if msg.Type == Create {
		create(msg.Data)
	}
}

func update(msg Data) {

}

func delete(msg Data) {

}

func create(msg Data) {

}
