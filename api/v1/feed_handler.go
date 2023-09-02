package v1

import (
	"douyin/database"
	"douyin/database/models"
	"log"

	"douyin/cache"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

type FeedRequest struct {
	LatestTime int64  `json:"latest_time,omitempty"` // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `json:"token,omitempty"`       // 用户登录状态下设置
}

type feedResponse struct {
	StatusCode int64 `json:"status_code"` // 状态码，0-成功，其他值-失败
	// StatusMsg  string `json:"status_msg"` // 返回状态描述
	NextTime  int64            `json:"next_time"`  // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
	VideoList []Video_feedResp `json:"video_list"` // 视频列表
}

// Video
type Video_feedResp struct {
	Author        Author_feedResp `json:"author"`         // 视频作者信息
	CommentCount  int64           `json:"comment_count"`  // 视频的评论总数
	CoverURL      string          `json:"cover_url"`      // 视频封面地址
	FavoriteCount int64           `json:"favorite_count"` // 视频的点赞总数
	ID            int64           `json:"id"`             // 视频唯一标识
	IsFavorite    bool            `json:"is_favorite"`    // true-已点赞，false-未点赞
	PlayURL       string          `json:"play_url"`       // 视频播放地址
	Title         string          `json:"title"`          // 视频标题
}

// 视频作者信息
//
// User
type Author_feedResp struct {
	Avatar          string `json:"avatar"`           // 用户头像
	BackgroundImage string `json:"background_image"` // 用户个人页顶部大图
	FavoriteCount   int64  `json:"favorite_count"`   // 喜欢数
	FollowCount     int64  `json:"follow_count"`     // 关注总数
	FollowerCount   int64  `json:"follower_count"`   // 粉丝总数
	ID              int64  `json:"id"`               // 用户id
	IsFollow        bool   `json:"is_follow"`        // true-已关注，false-未关注
	Name            string `json:"name"`             // 用户名称
	Signature       string `json:"signature"`        // 个人简介
	TotalFavorited  int64  `json:"total_favorited"`  // 获赞数量
	WorkCount       int64  `json:"work_count"`       // 作品数
}

// 处理获取用户feed流请求 /feed
func GetFeedHandler(c *gin.Context) {
	var request FeedRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 如果没有传递 latest_time 参数，则默认为当前时间
	if request.LatestTime == 0 {
		request.LatestTime = time.Now().Unix()
	}

	// 如果没有传递 token 参数，则默认用户为未登录状态
	var current_userID, _ = c.Get("user_id")

	var video_ids []int64
	var videos []models.Video

	// 根据时间戳倒序排序返回最多30个视频
	result := database.DB.Table("video").Where("create_time < ?", time.Unix(request.LatestTime, 0)).Order("create_time desc").Limit(30).Select("id").Find(&videos)
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
	for _, v_id := range video_ids {
		Videos = append(Videos, Get_Video_for_feed(v_id, int64(current_userID.(int))))
	}

	response := feedResponse{
		StatusCode: 0,                 // 成功状态码
		NextTime:   time.Now().Unix(), // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time

		VideoList: Videos, // 视频列表
	}

	c.JSON(http.StatusOK, response)

}

func Get_Video_for_feed(video_id int64, current_userID int64) Video_feedResp {
	var video models.Video
	result := database.DB.Table("video").Where("id = ?", video_id).Find(&video)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	autherID := video.AuthorUserID
	author_resp := Get_author_for_feed(autherID, current_userID)

	var far models.Favorite
	var isfar bool
	result2 := database.DB.Table("favorite").Where("user_id = ? AND video_id = ? AND is_deleted=-1", current_userID, video_id).First(&far)
	if result2.Error != nil && result2.Error != gorm.ErrRecordNotFound {
		log.Println(result2.Error)
	}

	if result2.RowsAffected > 0 {
		isfar = true
	} else {
		isfar = false
	}
	likes, _ := cache.GetVideoLikesFromRedis(video_id)
	var video_resp = Video_feedResp{
		ID:            video_id,
		Author:        author_resp,
		PlayURL:       video.PlayURL,
		CoverURL:      video.CoverURL,
		FavoriteCount: likes,
		CommentCount:  int64(video.Comments),
		IsFavorite:    isfar,
		Title:         video.Title,
	}

	return video_resp
}

func Get_author_for_feed(author_id int64, current_userID int64) Author_feedResp {

	var author_resp Author_feedResp
	var author models.User
	var relation models.Relation
	var follow bool

	result1 := database.DB.Table("user").Where("id = ?", author_id).First(&author)
	if result1.Error != nil {
		log.Println(result1.Error)
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

	FavoriteCount, _ := cache.GetFavoriteCountFromRedis(author_id)
	TotalFavorited, _ := cache.GetTotalFavoritedFromRedis(author_id)

	author_resp = Author_feedResp{
		ID:              author_id,
		Name:            author.Name,
		BackgroundImage: author.BackgroundImage, // 用户个人页顶部大图
		FavoriteCount:   FavoriteCount,          // 喜欢数
		FollowCount:     author.FollowCount,     // 关注总数
		FollowerCount:   author.FollowerCount,   // 粉丝总数
		Signature:       author.Signature,       // 个人简介
		TotalFavorited:  TotalFavorited,         // 获赞数量
		WorkCount:       author.WorkCount,       // 作品数
		Avatar:          author.Avatar,
		IsFollow:        follow,
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
