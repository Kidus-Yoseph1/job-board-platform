package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
)

const (
	// Time allowed to write a message to the peer/browser.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer/browser.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (in bytes).
	maxMessageSize = 512
)

// Client represents a single active, authenticated WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan NotificationPayload
	log    *logger.Logger
}

// NewClient initializes a new Client instance.
func NewClient(hub *Hub, conn *websocket.Conn, userID string, log *logger.Logger) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		send:   make(chan NotificationPayload, 256),
		log:    log,
	}
}

// ReadPump loops and reads incoming messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Warnw("unexpected websocket close error", "error", err, "userID", c.userID)
			}
			break
		}
	}
}

// WritePump pushes outbound messages from the Hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case payload, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(payload); err != nil {
				c.log.Errorw("failed to write websocket JSON message", "error", err, "userID", c.userID)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Errorw("failed to send websocket ping heartbeat", "error", err, "userID", c.userID)
				return
			}
		}
	}
}
