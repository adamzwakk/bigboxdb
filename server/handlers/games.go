package games

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func All(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{
		"games": "none",
	})
}