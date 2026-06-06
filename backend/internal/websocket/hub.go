package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

type Client struct {
	ID       uuid.UUID
	GameID   uuid.UUID
	PlayerID uuid.UUID
	Conn     *websocket.Conn
	Send     chan []byte
}

type Hub struct {
	clients    map[uuid.UUID]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Message struct {
	GameID   uuid.UUID
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
	PlayerID uuid.UUID       `json:"player_id,omitempty"`
}

var hub *Hub
var hubOnce sync.Once

func GetHub() *Hub {
	hubOnce.Do(func() {
		hub = &Hub{
			clients:    make(map[uuid.UUID]*Client),
			broadcast:  make(chan *Message, 256),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		go hub.run()
	})
	return hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client %s connected", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				log.Printf("Client %s disconnected", client.ID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if client.GameID == message.GameID {
					select {
					case client.Send <- h.serializeMessage(message):
					default:
						close(client.Send)
						delete(h.clients, client.ID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) serializeMessage(msg *Message) []byte {
	data, _ := json.Marshal(msg)
	return data
}

func (h *Hub) BroadcastToGame(gameID uuid.UUID, msgType string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	msg := &Message{
		GameID: gameID,
		Type:   msgType,
		Data:   jsonData,
	}
	h.broadcast <- msg
}

func (h *Hub) SendToClient(clientID uuid.UUID, msgType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clients[clientID]
	if !ok {
		return
	}

	jsonData, _ := json.Marshal(data)
	msg := &Message{
		Type: msgType,
		Data: jsonData,
	}

	select {
	case client.Send <- h.serializeMessage(msg):
	default:
		close(client.Send)
		delete(h.clients, client.ID)
	}
}

func (h *Hub) SendToPlayer(gameID uuid.UUID, playerID uuid.UUID, msgType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	jsonData, _ := json.Marshal(data)
	msg := &Message{
		GameID:   gameID,
		Type:     msgType,
		Data:     jsonData,
		PlayerID: playerID,
	}

	for _, client := range h.clients {
		if client.GameID == gameID && client.PlayerID == playerID {
			select {
			case client.Send <- h.serializeMessage(msg):
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}
