package handlers

import (
	"net/http"
	"fmt"
	"time"
	"strconv"

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
	Region		string	`json:"region"`
	Year		int		`json:"year"`
	Platform	string	`json:"platform"`
	W			float32	`json:"w"`
	H			float32	`json:"h"`	
	D			float32	`json:"d"`
	Direction	int		`json:"dir"`
	GatefoldTransparent	bool	`json:"gatefold_transparent"`
	BoxType		uint	`json:"box_type"`
	BoxTypeName		string	`json:"box_type_name"`
	Developer	string	`json:"developer,omitempty"`
	DeveloperID	uint	`json:"developer_id,omitempty"`
	Publisher	string	`json:"publisher,omitempty"`
	PublisherID	uint	`json:"publisher_id,omitempty"`
	TexturePath	string	`json:"textureFileName"`
	ContributedBy	string	`json:"contributed_by"`
	AddedOn		time.Time	`json:"created_at"`
}

type queryOptions struct {
	Order			string
	WhereId			int
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
	var variant = getVariants(o)[0]
	c.JSON(http.StatusOK, variant)
}

func VariantById(c *gin.Context){
	id, _ := strconv.Atoi(c.Param("id"))
	var o = queryOptions{WhereId:id, Limit:1, Offset:0} 
	var variant = getVariants(o)[0]
	c.JSON(http.StatusOK, variant)
}

func getVariants(options queryOptions) []VariantResponse{
	d := db.GetDB()

	var variants []models.Variant

	q := d.Joins("Game", d.Select("id", "Title", "Slug", "Year", "PlatformID")).
	Preload("BoxType", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	}).Preload("Region", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	}).Preload("Game.Platform", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	}).Preload("User", func(db *gorm.DB) *gorm.DB {
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

	if options.WhereId > 0 {
		q = q.Where("variants.id = ?", options.WhereId)
	}

	if options.Order != "" {
		q = q.Order(options.Order)
	}

	if options.Limit != 0 {
		q = q.Limit(options.Limit)
	}
	
	if options.Limit == 1 {
		q.First(&variants)
	} else {
		q.Find(&variants)
	}

	var resp []VariantResponse
	for _, v := range variants {
		dir := 0
		if v.BoxType.ID == models.FindBoxTypeIDByName("Eidos Trapezoid") {
			dir = 1
		}
		resp = append(resp, VariantResponse{
			ID:            v.ID,
			GameID:		v.Game.ID,
			GameTitle:	v.Game.Title,
			VariantDesc:	fmt.Sprintf("%s %s", v.Description, v.BoxType.Name),
			Slug:	fmt.Sprintf("%s/%d", v.Game.Slug, v.ID),
			Region:		v.Region.Name,
			Year:		v.Year,
			Platform:	v.Game.Platform.Name,
			DeveloperID: v.Developer.ID,
			Developer: v.Developer.Name,
			Publisher: v.Publisher.Name,
			PublisherID: v.Publisher.ID,
			GatefoldTransparent:	v.GatefoldTransparent,
			W:		v.Width,
			H:		v.Height,
			D:		v.Depth,
			Direction: dir,
			BoxType:	v.BoxType.ID,
			BoxTypeName:	v.BoxType.Name,
			TexturePath: fmt.Sprintf("/scans/%s/%d/%s", v.Game.Slug, v.ID, "box.glb"),
			ContributedBy: v.User.Name,
			AddedOn: v.CreatedAt,
		})
	}

	return resp
}