package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

var RedisPool *redis.Pool

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
}

func RedisMiddleware() gin.HandlerFunc {
	InitRedisPool() // 初始化 Redis 连接池
	if RedisPool != nil {
		fmt.Println("get")
	}
	return func(ctx *gin.Context) {
		ctx.Set("RedisPool", RedisPool) // 将连接池存入上下文
		ctx.Next()
	}
}

func CloseRedis() {
	RedisPool.Close()
}
