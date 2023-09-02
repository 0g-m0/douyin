package middleware

import (
	"douyin/config"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求的路径
		requestPath := c.Request.URL.Path

		// 如果是登录或注册接口，则跳过验证
		if requestPath == "/douyin/user/register/" || requestPath == "/douyin/user/login/" {
			c.Next()
			return
		}

		// 需要验证token的接口：视频流接口、用户信息接口、投稿接口、发布列表、赞操作、喜欢列表、评论操作、评论列表

		// 需要验证user_id的接口：用户信息接口、发布列表、喜欢列表，自己在接口中单独额外验证user_id与token中的user_id是否一致

		// 如果没有token，则设置user_id为0
		if c.Query("token") == "" && c.PostForm("token") == "" {
			c.Set("user_id", 0)
			c.Next()
			return
		}

		// 获取请求体中的Token
		tokenString := c.Query("token")
		if tokenString == "" {
			tokenString = c.PostForm("token") // 从 POST form-data 中获取 token
		}

		// 如果token为空，则设置user_id为0
		if tokenString == "" {
			c.Set("user_id", 0)
			c.Next()
			return
		}

		// 解析Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfigInstance.JWTSecretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
			fmt.Println("无效的Token")
			c.Abort()
			return
		}

		// 检查Token是否有效
		if token.Valid {
			// 检查Token是否已经过期
			if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token已过期"})
				c.Abort()
				return
			}

			// 提取用户标识
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userID, ok := claims["user_id"].(float64); ok {
					// 将用户ID存储在请求上下文中
					c.Set("user_id", int64(userID))
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token验证失败"})
		c.Abort()

	}
}
