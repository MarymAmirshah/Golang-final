package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"GoPower/db"
	"GoPower/repository"
	"GoPower/server"
	"GoPower/service"
	"GoPower/ws"
)

func main() {
	database, err := db.InitDB("gopower.db")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	stRepo := repository.NewStationRepository(database)
	csRepo := repository.NewConsumerRepository(database)
	gridService := service.NewGridService(database, stRepo, csRepo)

	hub := ws.NewWebSocketHub()
	go hub.Run()

	router := server.SetupRouter(gridService, stRepo, csRepo, hub)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	log.Println("GoPower server listening on :8080")

	// Wait for the operator to request a shutdown (Ctrl+C, or a process
	// manager sending SIGTERM).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received, draining in-flight requests...")

	// Stop accepting/serving new HTTP requests first, giving in-flight
	// ones a bounded window to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Only once the HTTP layer is down do we stop the WebSocket hub, so no
	// telemetry/alert broadcast is dropped mid-flight.
	hub.Stop()

	log.Println("shutdown complete")
}
