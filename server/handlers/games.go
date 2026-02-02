package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

func GamesAll(c *gin.Context){
	database := db.GetDB()

	var games []models.Game

	// database.Select("title", "age").Find(&games)
	database.Preload(clause.Associations).Find(&games)

	c.JSON(http.StatusOK, games)
}