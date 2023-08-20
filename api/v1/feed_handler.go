package v1

import (
	// "douyin/config"
	"douyin/database"
	"douyin/database/models"
	"log"
	// "github.com/jinzhu/gorm"
	// "os"
	// "strings"
	"github.com/gin-gonic/gin"
	// "strconv"
	"net/http"
	"time"
)

type feedRequest struct {
	LatestTime int64 `json:"latest_time,omitempty"`// 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `json:"token,omitempty"`      // 用户登录状态下设置
}
type feedResponse struct {
	
	StatusCode int64   `json:"status_code"`// 状态码，0-成功，其他值-失败
	// StatusMsg  string `json:"status_msg"` // 返回状态描述
	NextTime   int64  `json:"next_time"`  // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
	VideoList  []Video_feedResp `json:"video_list"` // 视频列表
}

// Video
type Video_feedResp struct {
	Author        Author_feedResp   `json:"author"`        // 视频作者信息
	CommentCount  int64  `json:"comment_count"` // 视频的评论总数
	CoverURL      string `json:"cover_url"`     // 视频封面地址
	FavoriteCount int64  `json:"favorite_count"`// 视频的点赞总数
	ID            int64  `json:"id"`            // 视频唯一标识
	IsFavorite    bool   `json:"is_favorite"`   // true-已点赞，false-未点赞
	PlayURL       string `json:"play_url"`      // 视频播放地址
	Title         string `json:"title"`         // 视频标题
}

// 视频作者信息
//
// User
type Author_feedResp struct {
	// Avatar          string `json:"avatar"`          // 用户头像
	// BackgroundImage string `json:"background_image"`// 用户个人页顶部大图
	// FavoriteCount   int64  `json:"favorite_count"`  // 喜欢数
	// FollowCount     int64  `json:"follow_count"`    // 关注总数
	// FollowerCount   int64  `json:"follower_count"`  // 粉丝总数
	ID              int64  `json:"id"`              // 用户id
	IsFollow        bool   `json:"is_follow"`       // true-已关注，false-未关注
	Name            string `json:"name"`            // 用户名称
	// Signature       string `json:"signature"`       // 个人简介
	// TotalFavorited  string `json:"total_favorited"` // 获赞数量
	// WorkCount       int64  `json:"work_count"`      // 作品数
}


//处理获取用户feed流请求 /feed
func GetFeedHandler(c *gin.Context) {
	var request feedRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	current_userID, _ := c.Get("user_id")

	var video_ids []int64
	var videos []models.Video
	result := database.DB.Table("video").Select("video_id").Find(&videos)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	for _, video := range videos {
		video_ids = append(video_ids, video.VideoID)
	}
	// fmt.Println(video_ids)

	//新发布的先刷到，将vid倒叙排列
	video_ids = reverseList(video_ids)
	var Videos []Video_feedResp
	for _,v_id := range video_ids {
		Videos = append(Videos,Get_Video_for_feed(v_id,current_userID.(int64)))
	}

	
	
	// timestamp := time.Now().Unix()
	response := feedResponse{
		StatusCode: 0, // 成功状态码
		NextTime   : time.Now().Unix(),  // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
	
		// StatusMsg  : "feed get success" ,// 返回状态描述
		VideoList  :Videos ,// 视频列表
	}

	c.JSON(http.StatusOK, response)

}

func Get_Video_for_feed(video_id int64,current_userID int64) Video_feedResp{
	var video models.Video
	result := database.DB.Table("video").Where("video_id = ?", video_id).Find(&video)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	autherID := video.AuthorUserID
	author_resp := Get_author_for_feed(autherID,current_userID)

	var far models.Favorite
	var isfar bool
	result2 := database.DB.Table("favorite").Where("user_id = ? AND video_id = ?", current_userID, video_id).First(&far)
	// if result2.Error != nil && result2.Error != gorm.ErrRecordNotFound {
	// 	log.Fatal(result2.Error)
	// }

	if result2.RowsAffected > 0 {
		isfar = true
	} else {
		isfar = false
	}

	var video_resp = Video_feedResp{
		ID:	video_id,
		Author:	author_resp,
		PlayURL: video.PlayURL,
		CoverURL: video.CoverURL,
		FavoriteCount: int64(video.Likes),
		CommentCount: int64(video.Comments),
		IsFavorite: isfar,
		Title: video.Title,

	}
	



	return video_resp
}

func Get_author_for_feed(author_id int64,current_userID int64) Author_feedResp{
	
	var author_resp Author_feedResp
	var author models.User
	var relation models.Relation
	var follow bool

	result1 := database.DB.Table("user").Where("id = ?", author_id).First(&author)
	if result1.Error != nil {
		log.Fatal(result1.Error)
	}

	result2 := database.DB.Table("relation").Where("follower_id = ? AND followed_id = ?", current_userID, author_id).First(&relation)
	// if result2.Error != nil && result2.Error != gorm.ErrRecordNotFound {
	// 	log.Fatal(result2.Error)
	// }

	if result2.RowsAffected > 0 {
		follow = true
	} else {
		follow = false
	}

	author_resp = Author_feedResp{
		ID:            author_id,
		Name:          author.Name,
		// FollowCount:   0,
		// FollowerCount: 0, 
		IsFollow:      follow,
	}

	return author_resp
}


func reverseList(list []int64) []int64 {
	length := len(list)
	reversed := make([]int64, length)
	for i, j := 0, length-1; i < length; i, j = i+1, j-1 {
		reversed[j] = list[i]
	}
	return reversed
}