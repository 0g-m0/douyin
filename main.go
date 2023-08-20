package main

import (
	v1 "douyin/api/v1"
	//"douyin/config"
	"douyin/database"
	"douyin/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	//config.LoadConfig()
	database.ConnectDB()
	middleware.RedisMiddleware()
	defer middleware.CloseRedis()
	router := gin.Default()

	router.Use(middleware.JWTMiddleware())

	v1Group := router.Group("/douyin")
	{
		v1.SetupRoutes(v1Group)
	}

	err := router.Run(":8080")
	if err != nil {
		return
	}
}
