package models

import (
	"time"
)

type Region struct{
	ID						uint
	Name					string	`gorm:"type:varchar(255);not null;"`
	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}