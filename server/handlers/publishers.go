package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type PubResponse struct {
	ID			uint	`json:"id"`
	Name		string	`json:"name"`
	Slug	string	`json:"slug"`
	VariantCount	uint	`json:"variant_count"`
}

func PublishersAll(c *gin.Context) {
    d := db.GetDB()
    var pubs []models.Publisher

    d.Debug().Select("publishers.*, COUNT(variants.id) as variant_count").
        Joins("LEFT JOIN variants ON publishers.id = variants.publisher_id").
        Group("publishers.id").
        Find(&pubs)

	var resp []PubResponse
	for _, pub := range pubs {
		resp = append(resp, PubResponse{
			ID:            pub.ID,
			Name:		pub.Name,
			Slug:	pub.Slug,
			VariantCount:	pub.VariantCount,
		})
	}

    c.JSON(http.StatusOK, resp)
}