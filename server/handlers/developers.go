package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type DevResponse struct {
	ID			uint	`json:"id"`
	Name		string	`json:"name"`
	Slug	string	`json:"slug"`
	VariantCount	uint	`json:"variant_count"`
}

func DevelopersAll(c *gin.Context) {
    d := db.GetDB()
    var devs []models.Developer

    d.Debug().Select("developers.*, COUNT(variants.id) as variant_count").
        Joins("LEFT JOIN variants ON developers.id = variants.developer_id").
        Group("developers.id").
        Find(&devs)

	var resp []DevResponse
	for _, dev := range devs {
		resp = append(resp, DevResponse{
			ID:            dev.ID,
			Name:		dev.Name,
			Slug:	dev.Slug,
			VariantCount:	dev.VariantCount,
		})
	}

    c.JSON(http.StatusOK, resp)
}