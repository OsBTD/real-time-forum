package websocket

import (
	"encoding/json"
	"log"
	"real-time-forum/internal/models"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	Broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.nickname)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.nickname)

		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}
func (h *Hub) GetOnlineUsers() []models.User {
	h.mu.Lock()
	defer h.mu.Unlock()

	users := make([]models.User, 0, len(h.clients))
	for client := range h.clients {
		users = append(users, models.User{
			ID:       client.userID,
			Nickname: client.nickname,
		})
	}
	return users
}

func (h *Hub) BroadcastOnlineUsers() {
	onlineUsers := h.GetOnlineUsers()
	msg := models.WebSocketMessage{
		Type:    "online_users",
		Payload: onlineUsers,
	}
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling online users: %v", err)
		return
	}
	h.Broadcast <- messageBytes
}
func (h *Hub) SendMessage(messageBytes []byte) {
	h.Broadcast <- messageBytes
}
