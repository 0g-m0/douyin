package models

import (
	"time"
)


type Relation struct {
	FollowedID  int       `gorm:"column:followed_id;primaryKey"`
	FollowerID  int       `gorm:"column:follower_id;primaryKey"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}