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

func f32(v float32) *float32 { return &v }

var BoxtypesEnum = []BoxType{
	{ID:1, Name: "Big Box"},
	{ID:2, Name: "Small Box"},
	{ID:3, Name: "Eidos Trapezoid", Width: f32(10), Height: f32(10), Depth: f32(2)},
	{ID:4, Name: "DVD Case Slipcover"},
	{ID:5, Name: "Old Small Box"},
	{ID:6, Name: "Box in Box"},
	{ID:7, Name: "Big Box With Gatefold"},
	{ID:8, Name: "Small Box With Gatefold"},
	{ID:9, Name: "Small Box With Vertical Gatefold"},
	{ID:10, Name: "Big Box With Back Gatefold"},
	{ID:11, Name: "New Small Box"},
	{ID:12, Name: "New Big Box"},
	{ID:13, Name: "Small Box For DVD"},
	{ID:14, Name: "Big Long Box"},
	{ID:15, Name: "Big Box With Vertical Gatefold But Horizontal"},
	{ID:16, Name: "Small Box With Gatefold Right Flap"},
	{ID:17, Name: "DVD Case Slipcover with Gatefold"},
	{ID:18, Name: "New Box in Box"},
	{ID:19, Name: "Vinyl Like With Gatefold"},
}

func FindBoxTypeIDByName(name string) (uint) {
    for _, item := range BoxtypesEnum {
        if item.Name == name {
            return item.ID
        }
    }
    return 0
}