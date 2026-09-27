package server

import (
	"net/http"

	"GoPower/handlers"
	"GoPower/repository"
	"GoPower/service"
	"GoPower/ws"

	"github.com/gin-gonic/gin"
)

func SetupRouter(srv *service.GridService, stRepo repository.StationRepository, csRepo repository.ConsumerRepository, hub *ws.WebSocketHub) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	router.GET("/ws", handlers.ServeWS(hub))

	v1 := router.Group("/api/v1")
	{
		stations := v1.Group("/stations")
		{
			stations.POST("", handlers.CreateStation(stRepo))
			stations.GET("", handlers.ListStations(stRepo))
			stations.GET("/:id", handlers.GetStationByID(stRepo))
		}

		consumers := v1.Group("/consumers")
		{
			consumers.POST("", handlers.CreateConsumer(csRepo))
			consumers.GET("", handlers.ListConsumers(csRepo))
			consumers.GET("/:id", handlers.GetConsumerByID(csRepo))
		}

		v1.POST("/generation", handlers.RecordGeneration(srv))
		v1.POST("/dispatch", handlers.DispatchEnergy(srv))

		v1.GET("/grid/summary", handlers.GetGridSummary(srv))
	}

	return router
}
