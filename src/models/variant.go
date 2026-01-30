package models

import (
	"time"
)

type Variant struct{
	ID						uint

	GameID					uint
	Game					Game	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	DeveloperID				uint
	Developer				Developer	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	PublisherID				uint
	Publisher				Publisher	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Description				string	`gorm:"type:varchar(255);"`
	Slug					string	`gorm:"type:varchar(255);not null;unique;"`
	Region					int


	BoxTypeID				uint
	BoxType					BoxType
	
	GatefoldTransparent		bool
	Width					float32 `gorm:"type:float"`
	Height					float32 `gorm:"type:float"`
	Depth					float32 `gorm:"type:float"`
	ScanNotes				*string	`gorm:"type:text;"`
	WorthFrontView			bool

	UserID					uint
	User					User	`gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}