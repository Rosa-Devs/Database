package manifest

import "encoding/json"

// Basic data about database
//
// Name
// PubSub Room
type Manifest struct {
	Name     string `json:"name"`
	PubSub   string `json:"pubsub"`
	Chiper   string `json:"chiper"`
	Optional string `json:"optional"`
}

func (a *Manifest) Serialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

// Deserialize deserializes a JSON string to an Action struct.
func (a *Manifest) Deserialize(jsonDaat []byte) error {
	err := json.Unmarshal(jsonDaat, a)
	if err != nil {
		return err
	}
	return nil
}
