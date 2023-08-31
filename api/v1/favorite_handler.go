package v1

import (
	"douyin/cache"
	//"database/sql"
	"douyin/database"
	"douyin/database/models"
	"fmt"

	//"log"
	//"time"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// const dbName string = "mysql"
// const dbConnect string = "root:123456@tcp(127.0.0.1:3306)/douyin?charset=utf8&parseTime=true" //设置数据库连接参数
const expireTime int = 2 * 24 * 60 * 60

func Response(ctx *gin.Context, httpStatus int, v interface{}) {
	ctx.JSON(httpStatus, v)
}

func ErrorResponse(ctx *gin.Context, statusCode int, errorMsg string) {
	Response(ctx, statusCode, gin.H{
		"status_code": statusCode,
		"status_msg":  errorMsg,
	})
}

func SuccessResponse(ctx *gin.Context) {
	Response(ctx, http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "success",
	})
}

// FavActionParams是获取点赞操作请求的结构体
type FavActionParams struct {
	Token      string `form:"token" binding:"required"`
	VideoId    int64  `form:"video_id" binding:"required"`
	ActionType int8   `form:"action_type" binding:"required,oneof=1 2"`
}

// FavListResponse是获取点赞列表响应的结构体
type FavListResponse struct {
	StatusCode int              `json:"status_code"`
	StatusMsg  string           `json:"status_msg"`
	VideoList  []models.VideoFA `json:"video_list"`
}

// 点赞视频
func FavoriteAction(ctx *gin.Context) {
	var favInfo FavActionParams
	err := ctx.ShouldBind(&favInfo)
	if err != nil {
		ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}
	tokenUids, _ := ctx.Get("user_id")
	tokenUid, _ := tokenUids.(int64)

	if err != nil {
		ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	redisPool := cache.RedisPool
	if redisPool != nil {
		fmt.Println("get")
	}
	err = FavoriteActionDo(tokenUid, favInfo.VideoId, favInfo.ActionType, redisPool)

	if err != nil {
		ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}
	SuccessResponse(ctx)
	// 获取 Redis 连接池

}

func FavoriteActionDo(uid, vid int64, action int8, redisPool *redis.Pool) error {
	//conn := redisPool.Get() //重用已有的连接
	//defer conn.Close()

	tx := database.DB.Begin() //开启数据库事务
	act := action == 1        // 设置 act 为 true 或 false

	err := FavoriteTableChange(tx, "favorite", uid, vid, act)
	if err != nil {
		tx.Rollback() //有错误回滚
		fmt.Println("Rollback the transaction")
		return fmt.Errorf("点赞失败")
	} else {
		tx.Commit()
		fmt.Println("Commit the transaction")
		CacheFavoriteAction(uid, vid, act, redisPool)
	}
	fmt.Println("All database operations completed.")
	return nil
}

func CacheFavoriteAction(uid, vid int64, action bool, redisPool *redis.Pool) error {
	conn := redisPool.Get() //重用已有的连接
	defer conn.Close()

	var authoruid int64
	// 获取视频对应的用户ID
	authoruid, err := cache.GetAuthorUserIdFromRedis(vid)
	if err != nil {
		fmt.Println("获取视频作者id错误", err)
		return err
	}

	if redisPool != nil {
		fmt.Println("Get conn!")
	}
	//使用 Redis 的哈希（Hash）来存储用户表和视频表的信息，使用集合（Set）来存储点赞关系。
	keyUser := "user:" + strconv.FormatInt(uid, 10)
	keyVideo := "video:" + strconv.FormatInt(vid, 10)
	keyAuthor := "user:" + strconv.FormatInt(authoruid, 10)
	var change int

	if action {
		conn.Send("LPUSH", "user:"+strconv.FormatInt(uid, 10)+":likes", vid) //存储点赞关系
		change = 1
	} else {
		conn.Send("LREM", "user:"+strconv.FormatInt(uid, 10)+":likes", 0, vid)
		change = -1
	}
	exist, _ := redis.Int(conn.Do("EXISTS", keyUser))
	if exist == 1 {
		conn.Send("HINCRBY", keyUser, "favorite_count", change)
		conn.Send("EXPIRE", keyUser, expireTime)
	} else {
		cache.GetFavoriteCountFromRedis(uid)
	}
	exist, _ = redis.Int(conn.Do("EXISTS", keyVideo))
	if exist == 1 {
		conn.Send("HINCRBY", keyVideo, "likes_count", change)
		conn.Send("EXPIRE", keyVideo, expireTime)
	} else {
		cache.GetVideoLikesFromRedis(vid)
	}
	exist, _ = redis.Int(conn.Do("EXISTS", keyAuthor))
	if exist == 1 {
		conn.Send("HINCRBY", keyAuthor, "total_favorited", change)
		conn.Do("EXPIRE", keyAuthor, expireTime)
	} else {
		cache.GetTotalFavoritedFromRedis(authoruid)
	}

	conn.Flush()

	return nil
}

// 更新favorite表，1代表取消点赞，-1代表未取消点赞
func FavoriteTableChange(db *gorm.DB, tableName string, userID int64, videoID int64, action bool) error {
	var fav models.Favorite
	authoruid, _ := cache.GetAuthorUserIdFromRedis(videoID)
	err := db.Table(tableName).Where("user_id = ? AND video_id = ?", userID, videoID).First(&fav).Error

	//now := time.Now()

	if action {
		if gorm.IsRecordNotFoundError(err) {
			// 插入新记录，is_deleted 为 0，created_at 和 update_at 为当前时间
			newFav := models.Favorite{
				UserID:       userID,
				VideoID:      videoID,
				AuthorUserID: authoruid,
			}
			err = db.Table(tableName).Create(&newFav).Error
			if err != nil {
				fmt.Println("插入记录失败:", err)
				return err
			}
			//fmt.Println("插入成功")
		} else {
			// 已点赞禁止再点赞
			return fmt.Errorf("操作错误！")
		}
	} else {
		if gorm.IsRecordNotFoundError(err) {
			//数据库中无点赞记录，取消点赞即非法操作
			if err != nil {
				fmt.Println("操作错误:", err)
				return err
			}
		} else {
			err = db.Delete(&fav).Error
			if err != nil {
				fmt.Println("更新记录失败:", err)
				return err
			}
			//fmt.Println("取消点赞更改记录成功")
		}
	}
	return nil
}

// 修改用户的 favorite_count 及 total_favorited
func ChangeUserFavoriteCount(db *gorm.DB, tableName string, userID int64, userORauther string, action bool) error {
	var changeValue int
	if action {
		changeValue = 1
	} else {
		changeValue = -1
	}

	//Sql := fmt.Sprintf("update %s set favorite_count = favorite_count + %d where id = %d", tableName, changeValue, userID)
	//_, err = db.Exec(Sql)
	//err := database.DB.Table(tableName).Where("id=?", userID).Update("favorite_count", gorm.Expr("favorite_count+?", changeValue)).Error
	err := db.Table(tableName).Where("id=?", userID).Update(userORauther, gorm.Expr(userORauther+" + ?", changeValue)).Error
	if err != nil {
		fmt.Println("更新 favorite_count 失败:", err)
		return err
	}
	return nil
	//defer db.Close()
}

// 修改视频的 likes_count
func ChangeVideoLikesCount(db *gorm.DB, tableName string, videoID int64, action bool) error {
	var changeValue int
	if action {
		changeValue = 1
	} else {
		changeValue = -1
	}

	//Sql := fmt.Sprintf("update %s set likes_count = likes_count + %d where video_id = %d", tableName, changeValue, videoID)
	//_, err = db.Exec(Sql)
	err := db.Table(tableName).Where("id=?", videoID).Update("likes", gorm.Expr("likes+?", changeValue)).Error
	if err != nil {
		fmt.Println("更新 likes_count 失败:", err)
		return err
	}
	return nil
	//defer db.Close()
}

// 获取点赞列表
func FavoriteList(ctx *gin.Context) {
	uID := ctx.Query("user_id")
	//token := ctx.Query("token")
	userID, _ := strconv.ParseInt(uID, 10, 64)

	videoId := GetVideoId(userID)
	//fmt.Println("videoId==", videoId)
	resp, err := GetVideoById(videoId)
	if err != nil {
		fmt.Println("出错了，获取视频失败")
		//ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取视频失败"})
		ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response := FavListResponse{
		StatusCode: 0, // 成功状态码
		StatusMsg:  "获取视频成功",
		VideoList:  resp,
	}

	ctx.JSON(http.StatusOK, response)
}

// 获取视频id
func GetVideoId(userID int64) []int64 {
	conn := cache.RedisPool.Get() //重用已有的连接
	defer conn.Close()
	key := "user:" + strconv.FormatInt(userID, 10) + ":likes"
	exist, _ := redis.Int(conn.Do("EXISTS", key))
	if exist != 1 { //缓存查询失败，将数据载入缓存
		var videoId []int64
		err := database.DB.Table("favorite").Where("user_id=?", userID).Order("id desc, video_id").Pluck("video_id", &videoId).Error
		if err != nil {
			fmt.Println(err)
			return nil
		}
		for _, id := range videoId {
			conn.Send("RPUSH", key, id)
		}
		conn.Flush()
		fmt.Println("load cache OK!")
		return videoId
	}
	len, err := redis.Int(conn.Do("LLEN", key))
	var videoId []int64
	//fmt.Println(userID)
	//err := database.DB.Table("favorite").Where("user_id=? AND is_deleted=-1", userID).Pluck("video_id", &videoId).Error
	for i := 0; i < len; i++ {
		var id int64
		id, err = redis.Int64(conn.Do("LINDEX", key, i))
		videoId = append(videoId, id)
	}
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(videoId)
	return videoId
}

// 获取视频实例
func GetVideoById(videoId []int64) ([]models.VideoFA, error) {
	var videos []models.VideoFA
	for _, id := range videoId {
		fmt.Println(id)
		var v models.Video
		var u models.User
		database.DB.Table("video").Where("id=?", id).Find(&v)
		//fmt.Println(v)
		database.DB.Table("user").Where("id=?", v.AuthorUserID).Find(&u)
		//fmt.Println(u)
		favorite_count, _ := cache.GetFavoriteCountFromRedis(u.ID)
		total_favorited, _ := cache.GetTotalFavoritedFromRedis(u.ID)
		likes_count, _ := cache.GetVideoLikesFromRedis(v.VideoID)
		userfa := models.UserFA{
			Avatar:          u.Avatar,
			BackgroundImage: u.BackgroundImage,
			FavoriteCount:   favorite_count,
			FollowCount:     u.FollowCount,
			FollowerCount:   u.FollowerCount,
			ID:              u.ID,
			IsFollow:        false,
			Name:            u.Name,
			Signature:       u.Signature,
			TotalFavorited:  strconv.FormatInt(total_favorited, 10),
			WorkCount:       u.WorkCount,
		}
		videofa := models.VideoFA{
			Author:        userfa,
			CommentCount:  int64(v.Comments),
			CoverURL:      v.CoverURL,
			FavoriteCount: likes_count,
			ID:            v.VideoID,
			IsFavorite:    true,
			PlayURL:       v.PlayURL,
			Title:         v.Title,
		}
		videos = append(videos, videofa)
	}
	return videos, nil
}
