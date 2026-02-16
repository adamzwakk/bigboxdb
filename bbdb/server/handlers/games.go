package handlers

import (
	"net/http"
	"fmt"
	// "log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"gorm.io/gorm"

	"github.com/adamzwakk/bigboxdb/server/db"
	"github.com/adamzwakk/bigboxdb/server/models"
)

type GameResponse struct {
	ID			uint	`json:"id"`
	Title		string	`json:"title"`
	Variant		[]MiniVariantResponse	`json:"variants"`
	Slug		string	`json:"slug"`
	Platform	string	`json:"platform"`
	Description	*string	`json:"description"`
	Links		[]LinkResponse `json:"links"`
	IgdbSlug		*string	`json:"igdb_slug,omitempty"`
	MobygamesID		*int	`json:"mobygames_id,omitempty"`
}

type MiniVariantResponse struct {
	ID			uint	`json:"id"`
	Desc		string	`json:"name"`
	BoxType		string	`json:"box_type_name"`
	TexturePath string	`json:"textureFileName"`
}

type LinkResponse struct {
	ID			uint	`json:"id"`
	Name		string	`json:"name"`
	Link		string	`json:"link"`
}

func GamesAll(c *gin.Context){
	database := db.GetDB()

	var games []models.Game

	database.Preload(clause.Associations).Find(&games)

	c.JSON(http.StatusOK, games)
}

func GameBySlug(c *gin.Context){
	slug := c.Param("slug")

	resp := getGames(queryOptions{WhereSlug: slug, Limit: 1})
	if resp != nil {
		c.JSON(http.StatusOK, resp[0])
	} else {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func getGames(options queryOptions) []GameResponse {
	d := db.GetDB()

	var games []models.Game

	q := d.Model(&models.Game{}).Preload("Variants", func(db *gorm.DB) *gorm.DB {
        return db.Select("id", "game_id", "description","box_type_id")
    }).Preload("Links", func(db *gorm.DB) *gorm.DB {
        return db.Select("id", "game_id", "type_id", "link")
    }).Preload("Links.Type", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "SmallName")
	}).Preload("Variants.BoxType", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
	}).Preload("Platform", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Name")
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

	if len(games) == 0 {
		return nil
	}

	var resp []GameResponse
	for _, g := range games {

		var miniVariants []MiniVariantResponse
		for _, v := range g.Variants {
			miniVariants = append(miniVariants, MiniVariantResponse{
				ID:	v.ID,
				Desc: v.Description,
				BoxType: v.BoxType.Name,
				TexturePath: fmt.Sprintf("/scans/%s/%d/%s", g.Slug, v.ID, "box.glb"),
			})
		}

		var links []LinkResponse
		for _, l := range g.Links {
			links = append(links, LinkResponse{
				ID:	l.ID,
				Name: l.Type.SmallName,
				Link: l.Link,
			})
		}

		resp = append(resp, GameResponse{
			ID:	g.ID,
			Title: g.Title,
			Slug: g.Slug,
			Platform: g.Platform.Name,
			Description: g.Description,
			MobygamesID: g.MobygamesID,
			IgdbSlug: g.IgdbSlug,
			Variant: miniVariants,
			Links: links,
		})
	}

	return resp
}