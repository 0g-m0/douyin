
package models

import (
	"time"
)
type Favorite struct {
	UserID     int       `gorm:"column:user_id;primaryKey"`
	VideoID    int       `gorm:"column:video_id;primaryKey"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}