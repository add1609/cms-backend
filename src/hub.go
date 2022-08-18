package src

import (
	"time"
)

type Hub struct {
	// Environment variables. Read in by main.go and then passed onto hub.go.
	env map[string]string

	// Time allowed to write to the Client.
	writeWait time.Duration

	// Time allowed to read the next pong message from the Client.
	pongWait time.Duration

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod time.Duration

	// Registered Client's.
	clients map[*Client]bool

	// Register requests from the Client's.
	register chan *Client

	// Unregister requests from Client's.
	unregister chan *Client
}

func NewHub(myEnv map[string]string) *Hub {
	return &Hub{
		env:        myEnv,
		writeWait:  10 * time.Second,
		pongWait:   60 * time.Second,
		pingPeriod: ((60 * time.Second) * 9) / 10,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// The $connect route
			h.clients[client] = true
			client.send <- []byte(client.url)
		case client := <-h.unregister:
			// The $disconnect route
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		}
	}
}
