package models

import (
	"time"
)

type Favorite struct {
	UserID    int64     `gorm:"column:user_id;primary_key;autoIncrement:false"`
	VideoID   int64     `gorm:"column:video_id;primary_key;autoIncrement:false"`
	CreatedAt time.Time `gorm:"column:created_at"`
}
