package manifest

// Basic data about database
//
// Name
// PubSub Room
type Manifest struct {
	Name   string `json:"name"`
	PubSub string `json:"pubsub"`
}
