package db

import (
	"sync"
)

const DbUpdateEvent = "DbUpdateEvent"

type Event struct {
	Name string
	Data []byte
}

type EventBus struct {
	lock      sync.Mutex
	listeners map[string][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]chan Event),
	}
}

func (bus *EventBus) Subscribe(event string, ch chan Event) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if _, exists := bus.listeners[event]; !exists {
		bus.listeners[event] = make([]chan Event, 0)
	}
	bus.listeners[event] = append(bus.listeners[event], ch)
}

func (bus *EventBus) Unsubscribe(event string, ch chan Event) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if listeners, exists := bus.listeners[event]; exists {
		for i, listener := range listeners {
			if listener == ch {
				bus.listeners[event] = append(bus.listeners[event][:i], bus.listeners[event][i+1:]...)
				close(ch)
				break
			}
		}
	}
}

func (bus *EventBus) Publish(event Event) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	if listeners, exists := bus.listeners[event.Name]; exists {
		for _, ch := range listeners {
			go func(ch chan Event) {
				ch <- event
			}(ch)
		}
	}
}

// func main() {
// 	// Create a new event bus
// 	eventBus := NewEventBus()

// 	// Subscribe to "DbChanged" event
// 	listener1 := make(chan Event)
// 	listener2 := make(chan Event)
// 	eventBus.Subscribe(DbChangeEvent, listener1)
// 	eventBus.Subscribe(DbChangeEvent2, listener2)

// 	// Set up a listener goroutine
// 	// Set up a listener goroutine
// 	go func() {
// 		for {
// 			for event := range listener1 {
// 				fmt.Println("Listener1: event:", event.Name)
// 			}
// 		}
// 	}()

// 	go func() {
// 		for {
// 			for event := range listener2 {
// 				fmt.Println("Listener2: event:", event.Name)
// 			}
// 		}
// 	}()

// 	// Simulate a DbChangeEvent
// 	go func() {
// 		for {
// 			time.Sleep(2 * time.Second)
// 			eventBus.Publish(Event{Name: DbChangeEvent, Data: []byte("Updated data")})
// 		}
// 	}()

// 	go func() {
// 		for {
// 			time.Sleep(2 * time.Second)
// 			eventBus.Publish(Event{Name: DbChangeEvent2, Data: []byte("Updated data")})
// 		}
// 	}()

// 	// Wait for goroutines to finish before exiting
// 	for {

// 	}
// }
