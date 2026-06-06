package websocket

import (
	"github.com/google/uuid"
)

type NotificationPayload struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Hub maintains the set of active WebSocket clients and handles the
// thread-safe routing/broadcasting of notifications to specific users.
type Hub struct {
	// clients maps a userID (as a string) to a set of active connections (*Client).
	clients map[string]map[*Client]bool

	// register channel receives incoming clients that want to establish a persistent connection.
	register chan *Client

	// unregister channel receives clients that disconnected or timed out.
	unregister chan *Client

	// broadcast channel receives messages meant for a SPECIFIC user ID.
	broadcast chan UserMessage
}

type UserMessage struct {
	UserID  string
	Payload NotificationPayload
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan UserMessage),
	}
}

// Run is the central event loop of the Hub.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true

		case client := <-h.unregister:
			if connections, ok := h.clients[client.userID]; ok {
				if _, exists := connections[client]; exists {
					delete(connections, client)
					close(client.send)

					if len(connections) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}

		case userMsg := <-h.broadcast:
			connections, ok := h.clients[userMsg.UserID]
			if ok {
				for client := range connections {
					select {
					case client.send <- userMsg.Payload:
					default:
						close(client.send)
						delete(connections, client)
						if len(connections) == 0 {
							delete(h.clients, client.userID)
						}
					}
				}
			}
		}
	}
}

// RegisterClient safely adds a newly connected client to the Hub.
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient safely removes a disconnected client from the Hub.
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// SendNotification routes a notification message to a specific user's active WebSocket connection(s).
func (h *Hub) SendNotification(userID uuid.UUID, payload NotificationPayload) {
	h.broadcast <- UserMessage{
		UserID:  userID.String(),
		Payload: payload,
	}
}
