package games

import (
	"net/http"
	"os"
	"log"
	"github.com/gin-gonic/gin"

	"github.com/adamzwakk/bigboxdb/server/db"
	"github.com/adamzwakk/bigboxdb/server/models"
)

func All(c *gin.Context){
	database := db.GetDB()

	if os.Getenv("APP_ENV") != "production" {
		database.AutoMigrate(&models.Game{})
	} else {
		log.Println("Skipping automatic migration in production.")
	}

	var games []models.Game

	// database.Select("title", "age").Find(&games)
	database.Find(&games)

	c.JSON(http.StatusOK, games)
}