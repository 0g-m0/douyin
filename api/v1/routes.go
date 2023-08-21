// api/v1/routes.go
package v1

import "github.com/gin-gonic/gin"

func SetupRoutes(router *gin.RouterGroup) {
	feedGroup := router.Group("/feed")
	{
		feedGroup.GET("/", GetFeedHandler)
	}
	userGroup := router.Group("/user")
	{
		userGroup.POST("/register/", UserRegisterHandler)
		userGroup.POST("/login/", UserLoginHandler)
		userGroup.GET("/", UserInfoHandler)
	}

	pubGroup := router.Group("/publish")
	{
		pubGroup.POST("/action/", UserPublishHandler)
		pubGroup.GET("/list/", PublishListHandler)
		
	}

	favoriteGroup := router.Group("/favorite")
	{
		favoriteGroup.POST("/action/", FavoriteAction)
		favoriteGroup.GET("/list/", FavoriteList)
	}
}
