package models

import (
	"time"
)

type BoxType struct{
	ID						uint
	Name					string	`gorm:"type:varchar(255);not null;"`
	Width					*float32 `gorm:"default:null"`
	Height					*float32 `gorm:"default:null"`
	Depth					*float32 `gorm:"default:null"`
	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}