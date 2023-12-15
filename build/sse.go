package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Serve a simple HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle SSE requests
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Set the headers for Server-Sent Events
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create a channel to send events to the client
		eventChan := make(chan string)

		// Register the channel for this client
		clients[r.RemoteAddr] = eventChan

		// Close the channel when the client disconnects
		defer func() {
			close(eventChan)
			delete(clients, r.RemoteAddr)
		}()

		// Loop to send events to the client
		for {
			select {
			case event := <-eventChan:
				// Send the event to the client
				fmt.Fprintf(w, "data: %s\n\n", event)
				w.(http.Flusher).Flush()
			case <-r.Context().Done():
				// Client disconnected
				return
			}
		}
	})

	// Start the HTTP server
	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

var clients = make(map[string]chan string)

func broadcastEvent(event string) {
	// Send the event to all connected clients
	for _, ch := range clients {
		go func(ch chan string) {
			ch <- event
		}(ch)
	}
}

func init() {
	// Simulate events being generated in the background
	go func() {
		for {
			time.Sleep(2 * time.Second)
			broadcastEvent("New event at " + time.Now().Format("2006-01-02 15:04:05"))
		}
	}()
}
