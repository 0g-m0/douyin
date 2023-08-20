package models

import "time"

type Comment struct {
	ID         int64      `gorm:"column:id;primary_key"`
	UserID     int64      `gorm:"column:user_id"`
	VideoID    int64      `gorm:"column:video_id"`
	Content    string     `gorm:"column:content"`
	CreateDate time.Time  `gorm:"column:create_date"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

// CommentDTO 是评论数据模型
type CommentDTO struct {
	ID         int64   `json:"id"`
	UserDTO    UserDTO `json:"user"`
	Content    string  `json:"content"`
	CreateDate string  `json:"create_date"`
}
