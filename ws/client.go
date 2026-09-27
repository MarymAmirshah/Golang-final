package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the
	// peer before the connection is considered dead.
	pongWait = 60 * time.Second

	// pingPeriod sends pings to the peer with this period; it must be less
	// than pongWait so a pong always has time to arrive before the deadline.
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize is the maximum message size accepted from the peer.
	// This hub does not process inbound client messages, so this is just a
	// safety limit against a misbehaving client.
	maxMessageSize = 4096
)

// Client represents a single active WebSocket connection registered with a
// WebSocketHub.
type Client struct {
	Hub  *WebSocketHub
	Conn *websocket.Conn
	Send chan []byte
	ID   string
}

// NewClient builds a Client wrapping an upgraded WebSocket connection.
func NewClient(hub *WebSocketHub, conn *websocket.Conn, id string) *Client {
	return &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		ID:   id,
	}
}

// ReadPump pumps messages from the WebSocket connection. It must run in its
// own goroutine for every client. This hub only broadcasts outward and does
// not act on inbound application messages, but a live read loop is still
// required so that control frames (pong, close) are processed and a dead
// or disconnected peer is detected; on any read error the client is
// unregistered from the hub and the connection is closed.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			return
		}
	}
}

// WritePump pumps messages from the client's Send channel to the WebSocket
// connection, and sends periodic ping heartbeats to keep the connection
// alive. It must run in its own goroutine for every client, and exits (and
// closes the connection) if a write fails or the hub closes Send.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed this client's channel.
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
