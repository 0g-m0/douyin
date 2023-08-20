// api/v1/routes.go
package v1

import "github.com/gin-gonic/gin"

func SetupRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/user")
	{
		userGroup.POST("/register/", UserRegisterHandler)
		userGroup.POST("/login/", UserLoginHandler)
		userGroup.GET("/", UserInfoHandler)
	}

	pubGroup := router.Group("/publish")
	{
		pubGroup.POST("/action/", UserPublishHandler)
	}

	favoriteGroup := router.Group("/favorite")
	{
		favoriteGroup.POST("/action/", FavoriteAction)
		favoriteGroup.GET("/list/", FavoriteList)
	}
}
