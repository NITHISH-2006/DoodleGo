package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Id         string
	Name       string
	RoomCode   string
	conn       *websocket.Conn
	send       chan *Message
	hub        *Hub
	registered chan struct{}
}

func NewClient(id, name, roomCode string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		Id:         id,
		Name:       name,
		RoomCode:   roomCode,
		conn:       conn,
		send:       make(chan *Message, 256),
		hub:        hub,
		registered: make(chan struct{}),
	}
}

func (c *Client) Serve() {
	go c.writePump()
	c.readPump()

	c.hub.unregister <- c
	c.conn.Close()
}

func (c *Client) readPump() {
	defer log.Printf("Client %s disconnected", c.Id)

	c.conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.hub.config.ReadTimeout))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error for %s: %v", c.Id, err)
			}
			break
		}

		msg.PlayerId = c.Id
		msg.Timestamp = time.Now().UnixMilli()

		switch msg.Type {
		case MsgStartGame:
			if r, ok := c.hub.GetRoom(c.RoomCode); ok {
				r.StartGame()
			}

		case MsgDrawStroke:
			var stroke StrokeData
			if err := json.Unmarshal(msg.Data, &stroke); err != nil {
				log.Printf("Invalid stroke data: %v", err)
				break
			}

			// Only the current drawer can draw
			if r, ok := c.hub.GetRoom(c.RoomCode); ok {
				r.mu.RLock()
				isDrawer := r.CurrentDrawer == c.Id && r.State == "playing"
				r.mu.RUnlock()

				if isDrawer {
					// Broadcast stroke to everyone in the room
					strokeMsg := newMessage(MsgStroke, stroke)
					strokeMsg.PlayerId = c.Id
					r.Broadcast(strokeMsg)
				}
			}

		case MsgGuess:
			var guess GuessData
			if err := json.Unmarshal(msg.Data, &guess); err != nil {
				log.Printf("Invalid guess data: %v", err)
				break
			}

			if r, ok := c.hub.GetRoom(c.RoomCode); ok {
				r.HandleGuess(c.Id, guess.Text)
			}

		default:
			log.Printf("Received from %s: %s", c.Name, msg.Type)
		}

	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.hub.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
