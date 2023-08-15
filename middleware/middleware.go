package middleware

import (
	"douyin/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求的路径
		requestPath := c.Request.URL.Path

		// 如果是登录或注册接口，则跳过验证
		if requestPath == "/douyin/user/register/" || requestPath == "/douyin/user/login" {
			c.Next()
			return
		}

		// 获取请求体中的Token
		var requestData map[string]interface{}
		if err := c.ShouldBind(&requestData); err != nil {
			// 处理请求体解析错误
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}

		tokenValue, ok := requestData["token"].(string)
		tokenValue2:=c.PostForm("token") 
		
		if  !ok || tokenValue == "" {
			// 处理没有 token 的情况
			if tokenValue2 == ""  {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
				fmt.Println(tokenValue)
				c.Abort()
				return

			}
			tokenValue = tokenValue2
			
		}

		token, err := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfigInstance.JWTSecretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
			c.Abort()
			return
		}

		// 将用户 ID 存储到上下文中，供后续处理使用
		claims, _ := token.Claims.(jwt.MapClaims)
		userID := int64(claims["user_id"].(float64))
		c.Set("user_id", userID)

		c.Next()
	}
}
