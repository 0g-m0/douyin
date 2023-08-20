package models

import (
	"time"
)

type Video struct {
	VideoID      int64      `gorm:"column:video_id;primaryKey"`
	AuthorUserID int64      `gorm:"column:author_user_id;not null"`
	PlayURL      string     `gorm:"column:play_url;size:256"`
	CoverURL     string     `gorm:"column:cover_url;size:256"`
	Likes        int        `gorm:"column:likes;default:0"`
	Comments     int        `gorm:"column:comments;default:0"`
	Title        string     `gorm:"column:title;size:50"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index"`
}
