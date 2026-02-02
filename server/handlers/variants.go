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
	ID			uint	`json:"variant_id"`
	GameID		uint	`json:"game_id"`
	GameTitle	string	`json:"title"`
	VariantDesc	string	`json:"variant"`
	Slug		string	`json:"slug"`
	Year		int		`json:"year"`
	Platform	string	`json:"platform"`
	W			float32	`json:"w"`
	H			float32	`json:"h"`	
	D			float32	`json:"d"`
	Direction	int		`json:"dir"`
	WorthFrontView	bool	`json:"worth_front_view"`
	GatefoldTransparent	bool	`json:"gatefold_transparent"`
	BoxType		uint	`json:"box_type"`
	Developer	string	`json:"developer"`
	Publisher	string	`json:"publisher"`
	TexturePath	string	`json:"textureFileName"`
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
		dir := 0
		if v.BoxType.ID == 3 {
			dir = 1
		}
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
			Direction: dir,
			BoxType:	v.BoxType.ID,
			TexturePath: fmt.Sprintf("/scans/%s/%d/%s", v.Game.Slug, v.ID, "box.glb"),
		})
	}

	c.JSON(http.StatusOK, resp)
}