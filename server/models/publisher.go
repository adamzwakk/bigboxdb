package models

import (
	"time"
)

type Publisher struct{
	ID						uint
	Name					string	`gorm:"type:varchar(255);not null;"`
	Slug					string	`gorm:"type:varchar(255);not null;unique;"`
	VariantCount 			uint `gorm:"->" json:"variant_count"`
	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}