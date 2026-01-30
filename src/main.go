package main

import (
	"log"
	"os"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	
	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
	"github.com/adamzwakk/bigboxdb-server/handlers"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load("./../.env"); err != nil {
			log.Println("No .env file found (ok for production)")
		}
	}

	// SEED/MIGRATE DB
	database := db.GetDB()
	if err := database.AutoMigrate(
        &models.Game{},
        &models.Variant{},
        &db.SeedMeta{},
    ); err != nil {
        log.Fatal(err)
    }
	if err := db.RunAllSeeds(database); err != nil {
        log.Fatal(err)
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
