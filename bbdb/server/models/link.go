package models

import (
	"time"
)

type Link struct{
	ID						uint

	TypeID					uint
	Type					LinkType

	GameID					uint
	Game					Game

	Link					string
	CreatedAt 				time.Time
	UpdatedAt 				time.Time
}