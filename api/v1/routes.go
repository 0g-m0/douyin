// api/v1/routes.go
package v1

import "github.com/gin-gonic/gin"

func SetupRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/user")
	{
		authGroup.POST("/register/", UserRegisterHandler)
		authGroup.POST("/login", UserLoginHandler)

		// 在这里添加对应的路由
	}
}
