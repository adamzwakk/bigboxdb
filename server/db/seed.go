package db

import (
	"log"
	"os"
	"fmt"
    "gorm.io/gorm"
	"github.com/dchest/uniuri"
	"github.com/adamzwakk/bigboxdb-server/models"
)

func RunAllSeeds(db *gorm.DB) error {
	seedsRan := 0
    if err := RunSeedOnce(db, "seed_v1_box_types", func(tx *gorm.DB) error {
		log.Println("No record of seed_v1_box_types seed, running...")
		seedsRan += 1
        return seedInitialBoxTypes(tx)
    }); err != nil {
        return err
    }

	if err := RunSeedOnce(db, "seed_v1_users", func(tx *gorm.DB) error {
		log.Println("No record of seed_v1_users seed, running...")
		seedsRan += 1
        return seedInitialUsers(tx)
    }); err != nil {
        return err
    }

	if err := RunSeedOnce(db, "seed_v1_link_types", func(tx *gorm.DB) error {
		log.Println("No record of seed_v1_link_types seed, running...")
		seedsRan += 1
        return seedInitialLinkTypes(tx)
    }); err != nil {
        return err
    }

	if(seedsRan > 0){
		log.Println(fmt.Sprintf("%d Seeds ran!", seedsRan))
	}

	return nil
}

func seedInitialBoxTypes(db *gorm.DB) error {
    for _, bt := range models.BoxtypesEnum {
        if err := db.
            Where("id = ?", bt.ID).
            FirstOrCreate(&bt).Error; err != nil {
				return err
			}
    }
    return nil
}

func seedInitialUsers(db *gorm.DB) error {
    users := []models.User{
        {Name: os.Getenv("BBDB_ADMIN_NAME"), ApiKey: uniuri.NewLen(24)},
    }

    for _, u := range users {
        if err := db.
            Where("id = ?", u.ID).
            FirstOrCreate(&u).Error; err != nil {
				return err
			}
    }
    return nil
}

func seedInitialLinkTypes(db *gorm.DB) error {
    for _, bt := range models.LinktypesEnum {
        if err := db.
            Where("id = ?", bt.ID).
            FirstOrCreate(&bt).Error; err != nil {
				return err
			}
    }
    return nil
}