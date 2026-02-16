package models

import (
	"time"
)

type Developer struct{
	ID						uint
	Name					string	`gorm:"type:varchar(255);not null;" json:"name"`
	Slug					string	`gorm:"type:varchar(255);not null;unique;" json:"slug"`
	VariantCount 			uint `gorm:"->" json:"variant_count"`
	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}