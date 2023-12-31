package cache

import (
	"fmt"
	"time"

	"douyin/database"
	"douyin/database/models"

	//"douyin/database/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	"log"

	"github.com/jinzhu/gorm"
)

var RedisPool *redis.Pool

const expireTime int = 2 * 24 * 60 * 60

func InitRedisPool() {
	// 初始化 Redis 连接池
	RedisPool = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   0,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				return nil, err
			}
			/*if _, err := c.Do("AUTH", "123456"); err != nil {
				c.Close()
				return nil, err
			}*/
			return c, err
		},
	}
	RedisPool.Get().Do("flushdb")
}

func RedisMiddleware() gin.HandlerFunc {
	InitRedisPool() // 初始化 Redis 连接池
	if RedisPool != nil {
		fmt.Println("Get Redis!")
	}
	LoadMysqlToRedis()
	return func(ctx *gin.Context) {
		ctx.Set("RedisPool", RedisPool) // 将连接池存入上下文
		ctx.Next()
	}
}

func CloseRedis() {
	RedisPool.Close()
}

type UserRedis struct {
	ID             int64
	TotalFavorited int64
	FavoriteCount  int64
}

type VideoRedis struct {
	ID           int64
	AuthorUserID int64
	Likes        int
}

type FavoriteRedis struct {
	UserID  int64
	VideoID int64
}

func LoadMysqlToRedis() {
	//load user
	conn := RedisPool.Get() //重用已有的连接
	defer conn.Close()
	/*var user []UserRedis
	err := database.DB.Table("user").Select([]string{"id", "total_favorited", "favorite_count"}).Scan(&user).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, u := range user {
		conn.Send("HMSET", "user:"+strconv.FormatInt(u.ID, 10), "total_favorited", u.TotalFavorited, "favorite_count", u.FavoriteCount)
	}
	//load video
	var video []VideoRedis
	err = database.DB.Table("video").Select([]string{"id", "author_user_id", "likes"}).Scan(&video).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range video {
		conn.Send("HMSET", "video:"+strconv.FormatInt(v.ID, 10), "author_user_id", v.AuthorUserID, "likes_count", v.Likes)
	}
	//load favorite
	var favorite []FavoriteRedis
	err := database.DB.Table("favorite").Select([]string{"user_id", "video_id"}).Order("id desc, video_id").Scan(&favorite).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, f := range favorite {
		conn.Send("RPUSH", "user:"+strconv.FormatInt(f.UserID, 10)+":likes", f.VideoID)
	}
	conn.Flush()
	fmt.Println("load cache OK!")*/
}

func GetFavoriteCountFromRedis(uid int64) (int64, error) {
	conn := RedisPool.Get()
	defer conn.Close()
	key := "user:" + strconv.FormatInt(uid, 10)
	res, err := redis.Int64(conn.Do("HGET", key, "favorite_count"))
	if err != nil {
		err = database.DB.Table("favorite").Where("user_id=?", uid).Count(&res).Error
		conn.Do("HSET", key, "favorite_count", res)
	}
	conn.Do("EXPIRE", key, expireTime)
	return res, err
}

func GetTotalFavoritedFromRedis(uid int64) (int64, error) {
	conn := RedisPool.Get()
	defer conn.Close()
	key := "user:" + strconv.FormatInt(uid, 10)
	res, err := redis.Int64(conn.Do("HGET", key, "total_favorited"))
	if err != nil {
		err = database.DB.Table("favorite").Where("author_user_id=?", uid).Count(&res).Error
		conn.Do("HSET", key, "total_favorited", res)
	}
	conn.Do("EXPIRE", key, expireTime)
	return res, err
}

func GetVideoLikesFromRedis(vid int64) (int64, error) {
	conn := RedisPool.Get()
	defer conn.Close()
	key := "video:" + strconv.FormatInt(vid, 10)
	res, err := redis.Int64(conn.Do("HGET", key, "likes_count"))
	if err != nil {
		err = database.DB.Table("favorite").Where("video_id=?", vid).Count(&res).Error
		conn.Do("HSET", key, "likes_count", res)
	}
	conn.Do("EXPIRE", key, expireTime)
	return res, err
}

func GetAuthorUserIdFromRedis(vid int64) (int64, error) {
	conn := RedisPool.Get()
	defer conn.Close()
	key := "video:" + strconv.FormatInt(vid, 10)
	res, err := redis.Int64(conn.Do("HGET", key, "author_user_id"))
	if err != nil {
		var v models.Video
		err = database.DB.Table("video").Where("id=?", vid).Select("author_user_id").Scan(&v).Error
		res = v.AuthorUserID
		conn.Do("HSET", key, "author_user_id", res)
	}
	conn.Do("EXPIRE", key, expireTime)
	return res, err
}

func CheckUserLikedVideo(uid, vid int64) (bool, error) {
	conn := RedisPool.Get()
	defer conn.Close()

	likeKey := "user:" + strconv.FormatInt(uid, 10) + ":likes"
	//vidStr := strconv.FormatInt(vid, 10)

	// 在redis缓存中查询
	exist, _ := redis.Int(conn.Do("EXISTS", likeKey))
	if exist == 1 { //缓存存在，查询缓存
		len, _ := redis.Int(conn.Do("LLEN", likeKey))
		for i := 0; i < len; i++ {
			var id int64
			id, _ = redis.Int64(conn.Do("LINDEX", likeKey, i)) //遍历喜欢列表
			if id == vid {                                     //存在点赞关系
				return true, nil
			}
		}
		return false, nil
	}

	// 在 MySQL 数据库中查找
	var far models.Favorite
	var isfar bool
	result := database.DB.Table("favorite").Where("user_id = ? AND video_id = ?", uid, vid).First(&far)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Println(result.Error)
		return false, result.Error
	}

	if result.RowsAffected > 0 {
		isfar = true
	} else {
		isfar = false
	}

	return isfar, nil
}
