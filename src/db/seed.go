package db

import (
	"log"
    "gorm.io/gorm"
	"github.com/adamzwakk/bigboxdb-server/models"
)

func RunAllSeeds(db *gorm.DB) error {
    return RunSeedOnce(db, "seed_v1_box_types", func(tx *gorm.DB) error {
		log.Println("No record of seed_v1_box_types seed, running...")
        if err := seedInitialBoxTypes(tx); err != nil {
            return err
        }
        return nil
    })
}

func seedInitialBoxTypes(db *gorm.DB) error {
    boxtypes := []models.BoxType{
        {ID:1, Name: "Big Box"},
		{ID:2, Name: "Small Box"},
		{ID:3, Name: "Eidos Trapezoid"},
		{ID:4, Name: "DVD Case Slipcover"},
		{ID:5, Name: "Old Small Box"},
		{ID:6, Name: "Box in Box"},
		{ID:7, Name: "Big Box With Gatefold"},
		{ID:8, Name: "Small Box With Gatefold"},
		{ID:9, Name: "Small Box With Vertical Gatefold"},
		{ID:10, Name: "Small Box With Back Gatefold"},
		{ID:11, Name: "New Small Box"},
		{ID:12, Name: "New Big Box"},
		{ID:13, Name: "Small Box For DVD"},
		{ID:14, Name: "Big Long Box"},
		{ID:15, Name: "Big Box With Vertical Gatefold But Horizontal"},
		{ID:16, Name: "Small Box With Gatefold Right Flap"},
		{ID:17, Name: "DVD Case Slipcover with Gatefold"},
		{ID:18, Name: "New Box in Box"},
		{ID:19, Name: "Vinyl Like With Gatefold"},
    }

    for _, bt := range boxtypes {
        if err := db.
            Where("id = ?", bt.ID).
            FirstOrCreate(&bt).Error; err != nil {
            return err
        }
    }
    return nil
}
