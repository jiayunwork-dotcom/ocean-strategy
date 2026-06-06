package handlers

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	ws "ocean-strategy/internal/websocket"
)

func WebSocketHandler(c *websocket.Conn) {
	gameIDStr := c.Params("game_id")
	playerIDStr := c.Params("player_id")

	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		log.Printf("Invalid game ID: %v", err)
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		log.Printf("Invalid player ID: %v", err)
		return
	}

	client := &ws.Client{
		ID:       uuid.New(),
		GameID:   gameID,
		PlayerID: playerID,
		Conn:     c,
		Send:     make(chan []byte, 256),
	}

	hub := ws.GetHub()
	hub.Register(client)

	defer func() {
		hub.Unregister(client)
	}()

	go writePump(client)
	readPump(client)
}

func readPump(client *ws.Client) {
	defer func() {
		ws.GetHub().Unregister(client)
		client.Conn.Close()
	}()

	for {
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var message ws.Message
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Printf("Invalid message: %v", err)
			continue
		}

		message.PlayerID = client.PlayerID
		message.GameID = client.GameID
	}
}

func writePump(client *ws.Client) {
	defer client.Conn.Close()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
	}
}
