package models

import (
	"time"
)

type User struct {
	ID              int64      `gorm:"column:id;primary_key"`
	Name            string     `gorm:"column:name"`
	FollowCount     int64      `gorm:"column:follow_count"`
	FollowerCount   int64      `gorm:"column:follower_count"`
	IsFollow        bool       `gorm:"column:is_follow"`
	Avatar          string     `gorm:"column:avatar"`
	BackgroundImage string     `gorm:"column:background_image"`
	Signature       string     `gorm:"column:signature"`
	TotalFavorited  int64      `gorm:"column:total_favorited"`
	WorkCount       int64      `gorm:"column:work_count"`
	FavoriteCount   int64      `gorm:"column:favorite_count"`
	Password        string     `gorm:"column:password"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
}

type UserDTO struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	FollowCount     int64  `json:"follow_count"`
	FollowerCount   int64  `json:"follower_count"`
	IsFollow        bool   `json:"is_follow"`
	Avatar          string `json:"avatar"`
	BackgroundImage string `json:"background_image"`
	Signature       string `json:"signature"`
	TotalFavorited  int64  `json:"total_favorited"`
	WorkCount       int64  `json:"work_count"`
	FavoriteCount   int64  `json:"favorite_count"`
}
