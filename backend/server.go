package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	var initMsg Message
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	err = conn.ReadJSON(&initMsg)
	if err != nil {
		log.Printf("Initial message error %v", err)
		conn.Close()
		return
	}

	log.Printf("Received initial message: %+v", initMsg.Type)

	response := newMessage(MsgRoomJoined, map[string]string{
		"message": "Connnection established",
	})

	conn.WriteJSON(response)
	conn.Close()
}
