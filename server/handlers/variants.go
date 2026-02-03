package handlers

import (
	"net/http"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type VariantResponse struct {
	ID			uint	`json:"id"`
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
	Developer	string	`json:"developer,omitempty"`
	Publisher	string	`json:"publisher,omitempty"`
	TexturePath	string	`json:"textureFileName"`
	AddedOn		time.Time	`json:"created_at"`
}

type queryOptions struct {
	Order			string
	Limit			int
	Offset			int
	WithDeveloper	bool
	WithPublisher	bool
}

func VariantsAll(c *gin.Context){
	var o = queryOptions{Order:"Game.Title asc", Limit:0, Offset:0, WithDeveloper:true, WithPublisher:true} 
	var variants = getVariants(o)
	c.JSON(http.StatusOK, variants)
}

func VariantsLatest(c *gin.Context){
	var o = queryOptions{Order:"created_at desc", Limit:20, Offset:0} 
	var variants = getVariants(o)
	c.JSON(http.StatusOK, variants)
}

func VariantsRandom(c *gin.Context){
	var o = queryOptions{Order:"rand()", Limit:1, Offset:0} 
	var variants = getVariants(o)[0]
	c.JSON(http.StatusOK, variants)
}

func getVariants(options queryOptions) []VariantResponse{
	d := db.GetDB()

	var variants []models.Variant

	q := d.Joins("Game", d.Select("id", "Title", "Slug", "Year", "PlatformID")).
	Preload("BoxType", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	}).Preload("Game.Platform", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	})

	if options.WithDeveloper {
		q = q.Preload("Developer", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		})
	}

	if options.WithPublisher {
		q = q.Preload("Publisher", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		})
	}

	if options.Order != "" {
		q = q.Order(options.Order)
	}

	if options.Limit != 0 {
		q = q.Limit(options.Limit)
	}

	q.Find(&variants)

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
			VariantDesc:	fmt.Sprintf("%s %s", v.Description, v.BoxType.Name),
			Slug:	fmt.Sprintf("%s/%d", v.Game.Slug, v.ID),
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
			AddedOn: v.CreatedAt,
		})
	}

	return resp
}