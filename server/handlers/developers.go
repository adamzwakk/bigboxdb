package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

func DevelopersAll(c *gin.Context){
	database := db.GetDB()

	var devs []models.Developer

	// database.Select("title", "age").Find(&games)
	database.Preload(clause.Associations).Find(&devs)

	c.JSON(http.StatusOK, devs)
}