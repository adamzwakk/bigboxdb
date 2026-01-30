package db

import (
    "errors"
    "time"

    "gorm.io/gorm"
)

type SeedMeta struct {
    ID    uint      `gorm:"primaryKey"`
    Name  string    `gorm:"uniqueIndex"`
    RanAt time.Time
}

func RunSeedOnce(db *gorm.DB, name string, fn func(*gorm.DB) error) error {
    var meta SeedMeta

    err := db.Where("name = ?", name).First(&meta).Error
    if err == nil {
        return nil // already ran
    }

    if !errors.Is(err, gorm.ErrRecordNotFound) {
        return err
    }

    return db.Transaction(func(tx *gorm.DB) error {
        if err := fn(tx); err != nil {
            return err
        }

        return tx.Create(&SeedMeta{
            Name:  name,
            RanAt: time.Now(),
        }).Error
    })
}
