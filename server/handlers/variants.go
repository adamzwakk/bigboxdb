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
	Select			string
	Order			string
	WhereId			int
	Limit			int		`default:"0"`
	Offset			int		`default:"0"`
	GroupBy			string
	WithDeveloper	bool
	WithPublisher	bool
}

type BoxTypeCount struct {
    Name  string	`json:"name"`
    Count int64		`json:"count"`
}

func VariantsAll(c *gin.Context) {
    variants, err := db.GetOrSetCache("variants:all", 10*time.Minute, func() ([]VariantResponse, error) {
        o := queryOptions{Order: "Game.Title asc", WithDeveloper: true, WithPublisher: true}
        return getVariants(o), nil
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, variants)
}

func VariantsLatest(c *gin.Context) {
    variants, err := db.GetOrSetCache("variants:latest", 10*time.Minute, func() ([]VariantResponse, error) {
        o := queryOptions{Order: "created_at desc", Limit: 20}
        return getVariants(o), nil
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, variants)
}

func VariantById(c *gin.Context) {
    id := c.Param("id")
    key := fmt.Sprintf("variant:%s", id)
    
    variant, err := db.GetOrSetCache(key, 5*time.Minute, func() (VariantResponse, error) {
        idInt, _ := strconv.Atoi(id)
        o := queryOptions{WhereId: idInt, Limit: 1}
        return getVariants(o)[0], nil
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, variant)
}

func VariantsRandom(c *gin.Context) {
    // Don't cache random - always fetch fresh
    o := queryOptions{Order: "rand()", Limit: 1}
    c.JSON(http.StatusOK, getVariants(o)[0])
}

func VariantsCountBoxTypes(c *gin.Context) {
    results, err := db.GetOrSetCache("variants:count_by_box_type", 10*time.Minute, func() ([]BoxTypeCount, error) {
        d := db.GetDB()
        var results []BoxTypeCount
        err := d.Model(&models.Variant{}).
            Joins("JOIN box_types ON box_types.id = variants.box_type_id").
            Select("box_types.name, COUNT(*) as count").
            Group("box_types.name").
            Scan(&results).Error
        return results, err
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, results)
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

	if options.GroupBy != "" {
		q = q.Group(options.GroupBy)
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
			VariantDesc:	fmt.Sprintf("%s", v.Description),
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