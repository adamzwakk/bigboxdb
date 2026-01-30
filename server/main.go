package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/adamzwakk/bigboxdb/server/handlers"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found (ok for production)")
		}
	}

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
