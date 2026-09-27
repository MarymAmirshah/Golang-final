package handlers

import (
	"errors"
	"net/http"

	"GoPower/models"
	"GoPower/repository"

	"github.com/gin-gonic/gin"
)

func CreateConsumer(repo repository.ConsumerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var consumer models.ConsumerMeter
		if err := c.ShouldBindJSON(&consumer); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := repo.Create(&consumer); err != nil {
			if errors.Is(err, models.ErrDuplicateConsumerID) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "consumer created", "consumer": consumer})
	}
}

func GetConsumerByID(repo repository.ConsumerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		consumer, err := repo.GetByID(id)
		if err != nil {
			if errors.Is(err, models.ErrConsumerNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"consumer": consumer})
	}
}

func ListConsumers(repo repository.ConsumerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		consumers, err := repo.ListAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if consumers == nil {
			consumers = []models.ConsumerMeter{}
		}

		c.JSON(http.StatusOK, gin.H{"consumers": consumers})
	}
}
