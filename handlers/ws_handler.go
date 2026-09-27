package handlers

import (
	"GoPower/ws"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWS upgrades the incoming HTTP request to a WebSocket connection,
// registers a new Client with the hub, and starts its read/write pumps.
func ServeWS(hub *ws.WebSocketHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := ws.NewClient(hub, conn, c.Request.RemoteAddr)
		hub.RegisterClient(client)

		go client.WritePump()
		go client.ReadPump()
	}
}
