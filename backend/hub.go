package main

import (
	"log"
	"sync"
)

type Hub struct {
	config Config
	rooms  map[string]*Room

	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message

	mu sync.RWMutex
}

func NewHub(config Config) *Hub {
	return &Hub{
		config:     config,
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("Client %s registered", client.Id)
		case client := <-h.unregister:
			log.Printf("Client %s unregistered", client.Id)
		}
	}
}
