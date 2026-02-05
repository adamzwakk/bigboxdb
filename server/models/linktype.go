package models

import (
	"time"
)

type LinkType struct{
	ID						uint
	Name					string	`gorm:"type:varchar(255);not null;"`
	SmallName				string	`gorm:"type:varchar(255);not null;unique;"`
	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}

var LinktypesEnum = []LinkType{
	{Name: "Steam", SmallName:"steam"},
	{Name: "GOG", SmallName:"gog"},
	{Name: "Official Website", SmallName:"official"},
	{Name: "Other", SmallName:"other"},
}