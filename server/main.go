package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/adamzwakk/bigboxdb/server/handlers"
)

func main() {
	r := gin.Default()

	{
		a := r.Group("/api")

		a.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})

		g := a.Group("/games")
		g.GET("/all", games.All)
	}

	

	r.Run(":8080")
}
