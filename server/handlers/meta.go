package handlers

import (
    "encoding/json"
    "time"
	"fmt"

	"gorm.io/gorm"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type Meta struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Image       string `json:"image"`
}

func GetMeta(slug string) (Meta, bool) {
    // Check cache
    val, err := db.Rdb.Get(db.Ctx, "meta:"+slug).Result()
    if err == nil {
        var m Meta
        if json.Unmarshal([]byte(val), &m) == nil {
            return m, true
        }
    }

	d := db.GetDB()
    var v models.Variant
    if err := d.Where("slug = ?", slug).Preload("Game", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "Title", "Slug")
	}).First(&v).Error; err != nil {
        return Meta{}, false
    }

    // Build and cache
    m := Meta{
        Title:       v.Game.Title,
        Description: v.Description,
        Image:       fmt.Sprintf("/scans/%s/%d/%s", v.Game.Slug, v.ID, "front.webp"),
    }
    setMeta(slug, m)

    return m, true
}

func setMeta(slug string, m Meta) {
    data, _ := json.Marshal(m)
    db.Rdb.Set(db.Ctx, "meta:"+slug, data, 15*time.Minute)
}

func deleteMeta(slug string) {
    db.Rdb.Del(db.Ctx, "meta:"+slug)
}