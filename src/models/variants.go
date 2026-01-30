package models

import (
	"time"
)

type Variant struct{
	ID						uint
	GameID					uint
	Game					Game	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Description				string	`gorm:"type:varchar(255);"`
	Slug					string	`gorm:"type:varchar(255);not null;unique;"`
	Region					int
	BoxTypeId				int
	GatefoldTransparent		bool
	Width					float32 `gorm:"type:float"`
	Height					float32 `gorm:"type:float"`
	Depth					float32 `gorm:"type:float"`
	ScanNotes				*string	`gorm:"type:text;"`
	WorthFrontView			bool
	ContributedBy			uint
	CreatedAt 				time.Time
}