package v1

import (
	"douyin/database"
	"douyin/database/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

// CommentActionRequest 是评论操作请求的结构体
type CommentActionRequest struct {
	Token       string `form:"token" binding:"required"`
	VideoID     int64  `form:"video_id" binding:"required"`
	ActionType  int32  `form:"action_type" binding:"required"`
	CommentText string `form:"comment_text"`
	CommentID   int64  `form:"comment_id"`
}

// CommentActionResponse 是评论操作响应的结构体
type CommentActionResponse struct {
	StatusCode int32             `json:"status_code"`
	StatusMsg  string            `json:"status_msg"`
	CommentDTO models.CommentDTO `json:"comment,omitempty"`
}

// CommentListRequest 是获取评论列表请求的结构体
type CommentListRequest struct {
	Token   string `form:"token" binding:"required"`
	VideoID int64  `form:"video_id" binding:"required"`
}

// CommentListResponse 是获取评论列表响应的结构体
type CommentListResponse struct {
	StatusCode  int32               `json:"status_code"`
	StatusMsg   string              `json:"status_msg"`
	CommentList []models.CommentDTO `json:"comment_list"`
}

// CommentHandler 处理用户评论操作
func CommentActionHandler(c *gin.Context) {
	var request CommentActionRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(int64)
	fmt.Println(userID)

	// 根据 ActionType 判断是发布评论还是删除评论
	if request.ActionType == 1 {
		// 记录当前时间
		currentTime := time.Now()

		// 发布评论
		comment := models.Comment{
			UserID:     userID, // 假设这是评论用户的 ID
			Content:    request.CommentText,
			VideoID:    request.VideoID,
			CreateDate: currentTime, // 根据格式需求自行调整
		}
		// 开始数据库事务
		tx := database.DB.Begin()

		// 将评论数据保存到数据库
		if err := tx.Table("comment").Create(&comment).Error; err != nil {
			// 发生错误时回滚事务
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "评论发布失败"})
			return
		}

		// 更新视频表的评论数
		if err := tx.Table("video").Where("id = ?", request.VideoID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
			// 发生错误时回滚事务
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新视频评论数失败"})
			return
		}

		// 提交事务
		tx.Commit()

		// 获取评论用户的信息
		var userDTO models.UserDTO
		if err := database.DB.Table("user").Where("id = ?", userID).First(&userDTO).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}

		commentDTO := models.CommentDTO{
			ID:         comment.ID,
			UserDTO:    userDTO,
			Content:    request.CommentText,
			CreateDate: currentTime.Format("01-02"),
		}

		response := CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "评论发布成功",
			CommentDTO: commentDTO,
		}

		c.JSON(http.StatusOK, response)

	} else if request.ActionType == 2 {

		// 删除评论之前，查询用户id是否与评论用户id一致
		var comment models.Comment
		if err := database.DB.Table("comment").Where("id = ?", request.CommentID).First(&comment).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
			return
		}

		if comment.UserID != userID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户id不一致"})
			return
		}

		// 开始数据库事务
		tx := database.DB.Begin()

		// 删除评论, 使用deleted_at字段标记删除
		if err := tx.Table("comment").Where("id = ?", request.CommentID).Update("deleted_at", time.Now()).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "评论删除失败"})
			return
		}

		// 更新视频表的评论数
		if err := tx.Table("video").Where("id = ?", request.VideoID).UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error; err != nil {
			// 发生错误时回滚事务
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新视频评论数失败"})
			return
		}

		// 提交事务
		tx.Commit()

		response := CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "评论删除成功",
		}

		c.JSON(http.StatusOK, response)
	}
}

func CommentListHandler(c *gin.Context) {
	var request CommentListRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 根据VideoID查询评论列表
	var comments []models.Comment
	if err := database.DB.Table("comment").Where("video_id = ?", request.VideoID).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询评论失败"})
		return
	}

	// 遍历评论列表，获取每个评论的用户信息
	var commentDTOs []models.CommentDTO
	for _, comment := range comments {

		// 获取评论用户的信息
		var userDTO models.UserDTO
		if err := database.DB.Table("user").Where("id = ?", comment.UserID).First(&userDTO).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}

		commentDTO := models.CommentDTO{
			ID:         comment.ID,
			UserDTO:    userDTO,
			Content:    comment.Content,
			CreateDate: comment.CreateDate.Format("01-02"),
		}

		commentDTOs = append(commentDTOs, commentDTO)
	}

	// 返回评论列表
	response := CommentListResponse{
		StatusCode:  0,
		StatusMsg:   "获取评论列表成功",
		CommentList: commentDTOs,
	}

	c.JSON(http.StatusOK, response)

}
