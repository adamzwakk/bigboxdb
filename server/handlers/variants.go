package handlers

import (
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type VariantResponse struct {
	ID			uint
	GameID		uint
	GameTitle	string
	VariantDesc	string
	Slug		string
	Year		int
	Platform	string
	W			float32
	H			float32
	D			float32
	WorthFrontView	bool
	GatefoldTransparent	bool
	BoxType		uint
	Developer	string
	Publisher	string
}

func VariantsAll(c *gin.Context){
	d := db.GetDB()

	var variants []models.Variant

	d.Preload("Developer", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Publisher", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Preload("Game", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Title", "Slug", "Year", "PlatformID")
	}).Preload("BoxType", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	}).Preload("Game.Platform", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	}).Find(&variants)

	var resp []VariantResponse
	for _, v := range variants {
		resp = append(resp, VariantResponse{
			ID:            v.ID,
			GameID:		v.Game.ID,
			GameTitle:	v.Game.Title,
			VariantDesc:	v.Description,
			Slug:	fmt.Sprintf("/%s/%d", v.Game.Slug, v.ID),
			Year:		v.Game.Year,
			Platform:	v.Game.Platform.Name,
			Developer: v.Developer.Name,
			Publisher: v.Publisher.Name,
			GatefoldTransparent:	v.GatefoldTransparent,
			WorthFrontView:	v.WorthFrontView,
			W:		v.Width,
			H:		v.Height,
			D:		v.Depth,
			BoxType:	v.BoxType.ID,
		})
	}

	c.JSON(http.StatusOK, resp)
}