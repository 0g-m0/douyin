// api/v1/routes.go
package v1

import "github.com/gin-gonic/gin"

func SetupRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/user")
	{
		authGroup.POST("/register/", UserRegisterHandler_origin)
		authGroup.POST("/login", UserLoginHandler)

		// 在这里添加对应的路由
		authGroup.POST("/douyin/publish/action/", UserPublishHandler)
	}

	pubGroup := router.Group("/publish")
	{
		
		pubGroup.POST("/action/", UserPublishHandler)
	}
}
