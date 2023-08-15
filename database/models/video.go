package models

import (
	"time"
)

type Video struct {
	VideoID       int64      `gorm:"primaryKey"`
	AuthorUserID  int64      `gorm:"not null"`
	PlayURL       string     `gorm:"size:256"`
	CoverURL      string     `gorm:"size:256"`
	Likes         int        `gorm:"default:0"`
	Comments      int        `gorm:"default:0"`
	Title         string     `gorm:"size:50"`
	CreatedAt     time.Time  `gorm:"not null"`
	UpdatedAt     *time.Time `gorm:"autoUpdateTime"`
	DeletedAt     *time.Time `gorm:"index"`
}