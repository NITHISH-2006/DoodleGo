package main

import (
	"encoding/json"
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
	done     chan struct{}
}

func NewClient(id, name, roomCode string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		Id:       id,
		Name:     name,
		RoomCode: roomCode,
		conn:     conn,
		send:     make(chan *Message, 256),
		hub:      hub,
		done:     make(chan struct{}),
	}
}

func (c *Client) trySend(msg *Message) {
	select {
	case c.send <- msg:
	default:
		log.Printf("Client %s send buffer full, dropping %s", c.Id, msg.Type)
	}
}

func (c *Client) Serve() {
	go c.writePump()
	c.readPump()
	c.hub.Unregister(c)
	close(c.done)
	c.conn.Close()
}

func (c *Client) readPump() {
	defer log.Printf("Client %s disconnected", c.Id)

	c.conn.SetReadLimit(c.hub.config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.hub.config.ReadTimeout))
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

		room, roomExists := c.hub.GetRoom(c.RoomCode)
		if !roomExists {
			break
		}

		switch msg.Type {
		case MsgStartGame:
			room.StartGame()

		case MsgDrawStroke:
			var stroke StrokeData
			if err := json.Unmarshal(msg.Data, &stroke); err != nil {
				log.Printf("Invalid stroke data: %v", err)
				break
			}
			room.HandleStroke(c.Id, stroke)

		case MsgUndoStroke:
			room.HandleUndo(c.Id)

		case MsgClearCanvas:
			room.HandleClear(c.Id)

		case MsgGuess:
			var guess GuessData
			if err := json.Unmarshal(msg.Data, &guess); err != nil {
				log.Printf("Invalid guess data: %v", err)
				break
			}
			room.HandleGuess(c.Id, guess.Text)

		case MsgToggleMissionTask:
			var data MissionToggleData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Printf("Invalid mission task data: %v", err)
				break
			}
			room.HandleMissionToggle(data.Index)

		case MsgNextMission:
			room.HandleNextMission()

		case MsgStartRelay:
			room.StartRelay()

		case MsgLeaveRoom:
			c.hub.Unregister(c)
			return

		case MsgPong:
			// The pong handler above already refreshed the deadline.

		default:
			log.Printf("Received from %s: %s", c.Name, msg.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.hub.config.PingInterval)
	defer ticker.Stop()
	defer c.conn.Close()

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

		case <-c.done:
			return
		}
	}
}
