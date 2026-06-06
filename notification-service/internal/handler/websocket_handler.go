package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "github.com/kidus-yoseph1/job-board-platform/notification-service/internal/websocket"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/response"
)

// Configure the Gorilla WebSocket Upgrader.
// This handles the transition from standard HTTP request-response protocol
// to the persistent, full-duplex TCP WebSocket protocol.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin allows us to control CORS for incoming WebSocket handshakes.
	// Returning true allows any origin to connect, which is critical for local development
	// where the frontend (e.g. React/Vite) might be running on a different port.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebsocketHandler houses dependencies required to manage WebSocket handshakes.
type WebsocketHandler struct {
	hub *ws.Hub
	log *logger.Logger
}

// NewWebsocketHandler initializes a new WebsocketHandler instance.
func NewWebsocketHandler(hub *ws.Hub, log *logger.Logger) *WebsocketHandler {
	return &WebsocketHandler{
		hub: hub,
		log: log,
	}
}

// HandleWS upgrades a standard HTTP GET connection into a persistent WebSocket channel.
// It assumes that the route is guarded by AuthMiddleware, which populates the "user_id" claim.
func (h *WebsocketHandler) HandleWS(c *gin.Context) {
	// 1. Retrieve the authenticated user's ID from the Gin context (injected by AuthMiddleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		h.log.Warnw("websocket handshake rejected: user_id missing from request context")
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.log.Infow("initiating websocket handshake", "userID", userIDStr)

	// 2. Perform the HTTP connection upgrade
	// This writes HTTP 101 Switching Protocols headers back to the browser.
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Errorw("failed to upgrade HTTP connection to WebSocket", "error", err, "userID", userIDStr)
		return
	}

	// 3. Create a new Client representation for this connection
	client := ws.NewClient(h.hub, conn, userIDStr, h.log)

	// 4. Register the new client connection with our central switchboard Hub.
	// This sends a pointer to the client into the Hub's `register` channel.
	h.hub.RegisterClient(client)

	// 5. Spin up the dedicated writePump loop in a separate concurrent goroutine.
	// This goroutine will run asynchronously to push outgoing notifications from the Hub.
	go client.WritePump()

	// 6. Execute the readPump loop synchronously in the current thread.
	// This blocks the Gin handler from exiting, keeping this specific TCP channel open.
	// Once the browser closes the connection or an error occurs, ReadPump returns,
	// and we cleanly exit this handler context.
	client.ReadPump()
}
