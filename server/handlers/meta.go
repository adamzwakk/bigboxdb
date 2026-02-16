package handlers

import (
    "encoding/json"
    "time"
	"fmt"
	"os"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type Meta struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Image       string `json:"image"`
}

func GetMeta(slug string, variantID int) (Meta, bool) {
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
    q := d.Joins("Game", d.Select("id", "Title", "Slug")).Joins("BoxType")
    if variantID > 0 {
        q = q.Where("variants.id = ?", variantID)
    } else {
        q = q.Where("Game.slug = ?", slug)
    }

    if err := q.First(&v).Error; err != nil {
        return Meta{}, false
    }

	title := v.Game.Title
	if variantID > 0 {
		title = fmt.Sprintf("%s (%s)", v.Game.Title, v.BoxType.Name)
	}

    // Build and cache
    m := Meta{
        Title:       title,
        Description: v.Description,
        Image:       fmt.Sprintf("%s/scans/%s/%d/%s", os.Getenv("SITE_URL"), v.Game.Slug, v.ID, "front.webp"),
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