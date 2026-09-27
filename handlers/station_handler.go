package handlers

import (
	"errors"
	"net/http"

	"GoPower/models"
	"GoPower/repository"

	"github.com/gin-gonic/gin"
)

func CreateStation(repo repository.StationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var station models.PowerStation
		if err := c.ShouldBindJSON(&station); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := repo.Create(&station); err != nil {
			if errors.Is(err, models.ErrDuplicateStationID) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "station created", "station": station})
	}
}

func GetStationByID(repo repository.StationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		station, err := repo.GetByID(id)
		if err != nil {
			if errors.Is(err, models.ErrStationNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"station": station})
	}
}

func ListStations(repo repository.StationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			stations []models.PowerStation
			err      error
		)

		if stationType := c.Query("type"); stationType != "" {
			stations, err = repo.GetByType(models.StationType(stationType))
		} else {
			stations, err = repo.ListAll()
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if stations == nil {
			stations = []models.PowerStation{}
		}

		c.JSON(http.StatusOK, gin.H{"stations": stations})
	}
}
