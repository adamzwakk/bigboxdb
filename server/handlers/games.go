package handlers

import (
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"gorm.io/gorm"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type GameResponse struct {
	ID			uint	`json:"id"`
	Title		string	`json:"title"`
	Variant		[]MiniVariantResponse	`json:"variants"`
	Slug		string	`json:"slug"`
	Description	*string	`json:"description"`
}

type MiniVariantResponse struct {
	ID			uint	`json:"id"`
	Desc		string	`json:"name"`
	TexturePath string	`json:"textureFileName"`
}

func GamesAll(c *gin.Context){
	database := db.GetDB()

	var games []models.Game

	database.Preload(clause.Associations).Find(&games)

	c.JSON(http.StatusOK, games)
}

func GameBySlug(c *gin.Context){
	slug := c.Param("slug")

	resp := getGames(queryOptions{WhereSlug: slug, Limit: 1})[0]

	c.JSON(http.StatusOK, resp)
}

func getGames(options queryOptions) []GameResponse {
	d := db.GetDB()

	var games []models.Game

	q := d.Model(&models.Game{}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
        return db.Select("id", "game_id", "description")
    })


	if options.WhereId > 0 {
		q = q.Where("games.id = ?", options.WhereId)
	} else if options.WhereSlug != "" {
		q = q.Where("games.slug = ?", options.WhereSlug)
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
		q.First(&games)
	} else {
		q.Find(&games)
	}

	var resp []GameResponse
	for _, g := range games {

		var miniVariants []MiniVariantResponse
		for _, v := range g.Variants {
			miniVariants = append(miniVariants, MiniVariantResponse{
				ID:	v.ID,
				Desc: v.Description,
				TexturePath: fmt.Sprintf("/scans/%s/%d/%s", g.Slug, v.ID, "box.glb"),
			})
		}

		resp = append(resp, GameResponse{
			ID:	g.ID,
			Title: g.Title,
			Slug: g.Slug,
			Description: g.Description,
			Variant: miniVariants,
		})
	}

	return resp
}