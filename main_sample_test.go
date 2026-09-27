package main

import (
	"GoPower/db"
	"GoPower/models"
	"GoPower/repository"
	"GoPower/server"
	"GoPower/service"
	"GoPower/ws"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestSample_WebSocketConnectionAndAlertBroadcast(t *testing.T) {
	database, err := db.InitDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	stRepo := repository.NewStationRepository(database)
	csRepo := repository.NewConsumerRepository(database)
	gridService := service.NewGridService(database, stRepo, csRepo)

	hub := ws.NewWebSocketHub()
	go hub.Run()
	defer hub.Stop()

	router := server.SetupRouter(gridService, stRepo, csRepo, hub)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Connect WebSocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	// Broadcast alert from hub
	if err := hub.BroadcastAlert("CRITICAL", "Severe load deficit detected"); err != nil {
		t.Fatalf("failed to broadcast alert: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from websocket: %v", err)
	}

	var alert models.GridAlert
	if err := json.Unmarshal(msgBytes, &alert); err != nil {
		t.Fatalf("failed to unmarshal alert payload: %v", err)
	}

	if alert.Severity != "CRITICAL" || alert.Type != "ALERT" {
		t.Errorf("received alert content mismatch: %+v", alert)
	}
}
