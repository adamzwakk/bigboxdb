package models

import (
	"time"
)

type User struct{
	ID						uint
	Name					string	`gorm:"type:varchar(255);not null;"`
	ApiKey					string	`gorm:"type:varchar(255);not null;unique;"`
	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}