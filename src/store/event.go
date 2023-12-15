package db

import "sync"

const DbChangeEvent = "DbChanged"

type Event struct {
	Name string
	Data []byte
}

type EventBus struct {
	lock sync.Mutex
}

func main() {

	db := new(DB)
	db.Start("/")

	go func() {

	}()
}
