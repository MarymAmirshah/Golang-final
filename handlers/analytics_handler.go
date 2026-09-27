package handlers

import (
	"net/http"

	"GoPower/service"

	"github.com/gin-gonic/gin"
)

func GetGridSummary(srv *service.GridService) gin.HandlerFunc {
	return func(c *gin.Context) {
		summary, err := srv.GetGridSummary()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"summary": summary})
	}
}
