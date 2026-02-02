package models

import (
	"time"
)

type Platform struct{
	ID						uint
	Name					string	`gorm:"type:varchar(255);not null;"`
	Slug					string	`gorm:"type:varchar(255);not null;unique;"`
	CreatedAt 				time.Time
}