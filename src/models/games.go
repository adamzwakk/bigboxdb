package models

import (
	"time"
)

type Game struct{
	ID						uint
	Title					string	`gorm:"type:varchar(255);not null;"`
	Description				*string	`gorm:"type:text;"`
	SeriesSort				*string `gorm:"type:varchar(255);"`
	Slug					string	`gorm:"type:varchar(255);not null;unique;"`
	Year					int
	Variants				[]Variant
	PlatformId				int
	MobygamesId				int
	IgdbId					int
	SteamLink				*string
	GogLink					*string
	OtherLink				*string
	CreatedAt 				time.Time
}