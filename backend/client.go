package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Id       string
	Name     string
	RoomCode string
	conn     *websocket.Conn
	send     chan *Message
	hub      *Hub
}

func NewClient(id, name, roomCode string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		Id:       id,
		Name:     name,
		RoomCode: roomCode,
		conn:     conn,
		send:     make(chan *Message, 256),
		hub:      hub,
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

		log.Printf("Received from %s: %s", c.Name, msg.Type)
		c.send <- &msg // Echo for now
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
