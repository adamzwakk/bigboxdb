package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

func PublishersAll(c *gin.Context){
	database := db.GetDB()

	var pubs []models.Publisher

	// database.Select("title", "age").Find(&games)
	database.Preload(clause.Associations).Find(&pubs)

	c.JSON(http.StatusOK, pubs)
}