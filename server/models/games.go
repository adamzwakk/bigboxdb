package models

import (
	"time"
)

type Game struct{
	ID			uint
	Title		string
	Slug		string
	CreatedAt 	time.Time
}