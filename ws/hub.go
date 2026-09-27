package ws

import (
	"encoding/json"
	"sync"
	"time"

	"GoPower/models"
)

// WebSocketHub coordinates every connected Client concurrently: it owns the
// client registry and is the single point through which telemetry and
// alert messages are broadcast to all of them.
type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	quit       chan struct{}
	wg         sync.WaitGroup
	isRunning  bool
}

// NewWebSocketHub builds a hub ready to run. Call go hub.Run() exactly once,
// right after construction, so that RegisterClient/Broadcast calls issued
// before the event loop goroutine is actually scheduled are still
// buffered/queued rather than lost.
//
// wg.Add(1) is done here rather than at the top of Run itself: Run executes
// in its own goroutine (started with `go hub.Run()`), so an Add call inside
// it could otherwise race with a concurrent Stop's wg.Wait() if Stop runs
// before that goroutine gets scheduled. Doing it here, synchronously during
// construction, guarantees it happens-before any later Stop call.
func NewWebSocketHub() *WebSocketHub {
	h := &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		quit:       make(chan struct{}),
		isRunning:  true,
	}
	h.wg.Add(1)
	return h
}

// Run is the hub's event loop. It owns the clients map exclusively while
// running, and must be started with `go hub.Run()` exactly once. It returns
// once Stop closes the quit channel.
func (h *WebSocketHub) Run() {
	defer h.wg.Done()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// The client's buffer is full (a stuck/slow reader):
					// drop it instead of blocking the whole hub.
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()

		case <-h.quit:
			h.mu.Lock()
			for client := range h.clients {
				close(client.Send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return
		}
	}
}

// RegisterClient adds a client to the hub. It is safe to call concurrently
// from any goroutine (typically the HTTP handler that just upgraded the
// connection). If the hub has already been stopped, this returns without
// blocking instead of deadlocking on a channel nobody is reading anymore.
func (h *WebSocketHub) RegisterClient(client *Client) {
	select {
	case h.register <- client:
	case <-h.quit:
	}
}

// UnregisterClient removes a client from the hub, closing its Send channel.
// Like RegisterClient, it never blocks past the hub being stopped.
func (h *WebSocketHub) UnregisterClient(client *Client) {
	select {
	case h.unregister <- client:
	case <-h.quit:
	}
}

// Stop gracefully shuts the hub down: it stops accepting new broadcasts,
// signals the Run loop to exit via the quit channel, and blocks until that
// goroutine has fully finished. Safe to call more than once.
func (h *WebSocketHub) Stop() {
	h.mu.Lock()
	if !h.isRunning {
		h.mu.Unlock()
		return
	}
	h.isRunning = false
	h.mu.Unlock()

	close(h.quit)
	h.wg.Wait()
}

// Broadcast enqueues a raw message to be sent to every connected client.
// It returns models.ErrHubStopped once the hub has been stopped.
func (h *WebSocketHub) Broadcast(message []byte) error {
	h.mu.RLock()
	running := h.isRunning
	h.mu.RUnlock()

	if !running {
		return models.ErrHubStopped
	}

	select {
	case h.broadcast <- message:
		return nil
	case <-h.quit:
		return models.ErrHubStopped
	}
}

// BroadcastTelemetry serializes a grid summary as a TELEMETRY message and
// broadcasts it to every connected client.
func (h *WebSocketHub) BroadcastTelemetry(summary *models.GridSummary) error {
	payload := models.TelemetryMessage{
		Type:      "TELEMETRY",
		Summary:   summary,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return h.Broadcast(data)
}

// BroadcastAlert builds and broadcasts an ALERT message with the given
// severity ("INFO", "WARNING", "CRITICAL") and human-readable text.
func (h *WebSocketHub) BroadcastAlert(severity, message string) error {
	alert := models.GridAlert{
		Type:      "ALERT",
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	return h.Broadcast(data)
}

// ClientCount returns the number of currently connected clients.
func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
