package v1

import (
	"database/sql"
	"douyin/database/models"
	"douyin/service"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	//"github.com/gomodule/redigo/redis"
	"douyin/middleware"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	successCode = 0
	errorCode   = 1
)

const dbName string = "mysql"
const dbConnect string = "root:123456@tcp(127.0.0.1:3306)/douyin?charset=utf8&parseTime=true" //设置数据库连接参数

func Response(ctx *gin.Context, httpStatus int, v interface{}) {
	ctx.JSON(httpStatus, v)
}

type FavActionParams struct {
	Token      string `form:"token" binding:"required"`
	VideoId    int64  `form:"video_id" binding:"required"`
	ActionType int8   `form:"action_type" binding:"required,oneof=1 2"`
}

type FavListResponse struct {
	StatusCode int            `json:"status_code"`
	StatusMsg  string         `json:"status_msg"`
	VideoList  []models.Video `json:"video_list"`
}

// 点赞视频
func FavoriteAction(ctx *gin.Context) {
	var favInfo FavActionParams
	err := ctx.ShouldBind(&favInfo)
	if err != nil {
		Response(ctx, 400, gin.H{"error": err.Error()})
		return
	}
	tokenUids, _ := ctx.Get("user_id")
	tokenUid, _ := tokenUids.(int64)

	if err != nil {
		Response(ctx, 500, gin.H{"error": err.Error()})
		return
	}

	redisPool := middleware.RedisPool
	if redisPool != nil {
		fmt.Println("get")
	}
	err = service.FavoriteAction(tokenUid, favInfo.VideoId, favInfo.ActionType, redisPool)

	if err != nil {
		Response(ctx, 500, gin.H{"error": err.Error()})
		return
	}
	Response(ctx, 200, gin.H{"message": "success"})
	// 获取 Redis 连接池

}

// 获取点赞列表
func FavoriteList(ctx *gin.Context) {
	uID := ctx.Query("user_id")
	//token := ctx.Query("token")
	userID, _ := strconv.ParseInt(uID, 10, 64)

	// 验证token是否有效
	/*if err := validateToken(userID, token); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
		return
	}*/

	videoId, _ := GetVideoId(userID)
	fmt.Println("videoId==", videoId)
	resp, err := GetVideoById(videoId)
	if err != nil {
		fmt.Println("出错了，....")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取视频失败"})
	}

	response := FavListResponse{
		StatusCode: 0, // 成功状态码
		StatusMsg:  "获取视频成功",
		VideoList:  resp,
	}

	ctx.JSON(http.StatusOK, response)
}

// 获取视频id
func GetVideoId(userID int64) ([]int64, error) {
	var videoId []int64
	db, _ := sql.Open(dbName, dbConnect)
	sql := fmt.Sprintf("select video_id from favorite where user_id=%d", userID)
	rows, err := db.Query(sql)
	if err != nil {
		log.Println(err)
		return videoId, err
	}
	defer rows.Close()
	for rows.Next() {
		var vid int64
		// 获取各列的值，放到对应的地址中
		rows.Scan(&vid)
		videoId = append(videoId, vid)
	}
	defer db.Close()
	return videoId, err
}

// 获取视频实例
func GetVideoById(videoId []int64) ([]models.Video, error) {
	var videos []models.Video
	db, err := gorm.Open(dbName, dbConnect)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var v models.Video
	for _, id := range videoId {
		fmt.Println(id)
		db.Table("video").Where("video_id=?", id).Find(&v)
		fmt.Println(v)
		videos = append(videos, v)
	}
	return videos, nil
}
