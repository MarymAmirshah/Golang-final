package handlers

import (
	"errors"
	"net/http"

	"GoPower/models"
	"GoPower/service"

	"github.com/gin-gonic/gin"
)

type RecordGenerationRequest struct {
	StationID string  `json:"station_id"`
	OutputMW  float64 `json:"output_mw"`
}

type DispatchRequest struct {
	StationID  string  `json:"station_id"`
	ConsumerID string  `json:"consumer_id"`
	AmountMW   float64 `json:"amount_mw"`
}

func RecordGeneration(srv *service.GridService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RecordGenerationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := srv.RecordGeneration(req.StationID, req.OutputMW); err != nil {
			if errors.Is(err, models.ErrStationNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			// invalid amount, station inactive, or output exceeding capacity
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "generation recorded"})
	}
}

func DispatchEnergy(srv *service.GridService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DispatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		record, err := srv.DispatchEnergy(req.StationID, req.ConsumerID, req.AmountMW)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrStationNotFound), errors.Is(err, models.ErrConsumerNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case errors.Is(err, models.ErrInsufficientEnergy),
				errors.Is(err, models.ErrConsumerDisconnected),
				errors.Is(err, models.ErrStationInactive),
				errors.Is(err, models.ErrStationOverload):
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			default:
				// invalid amount or other input error
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "dispatch successful", "record": record})
	}
}
