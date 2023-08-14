// api/v1/auth_handler.go
package v1

import (
	"douyin/config"
	"douyin/database"
	"douyin/database/models"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// UserRegisterRequest 是用户注册请求的结构体
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,max=32"`
	Password string `json:"password" binding:"required,max=32"`
}

// UserRegisterResponse 是用户注册响应的结构体
type UserRegisterResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	UserID     int64  `json:"user_id"`
	Token      string `json:"token"`
}

// UserRegisterHandler 处理用户注册请求
func UserRegisterHandler_origin(c *gin.Context) {
	var request UserRegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashPassword, err := hashPassword(request.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户数据模型
	user := models.User{
		Username: request.Username,
		Password: hashPassword,
	}

	// 保存用户数据到数据库
	if err := database.DB.Table("user").Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户注册失败"})
		return
	}

	// 模拟生成用户 ID 和 Token
	token := generateJWTToken(user.ID)

	response := UserRegisterResponse{
		StatusCode: 0, // 成功状态码
		StatusMsg:  "注册成功",
		UserID:     user.ID,
		Token:      token,
	}

	c.JSON(http.StatusOK, response)
}

func UserRegisterHandler(c *gin.Context) {
	// codes below has confliction with the apk provided by BD,
	// and this will incur a 400 code.
	
	// var request UserRegisterRequest
	// if err := c.ShouldBindJSON(&request); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	username := c.Query("username")
	password := c.Query("password")

	hashPassword, err := hashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户数据模型
	user := models.User{
		Username: username,
		Password: hashPassword,
	}

	// 保存用户数据到数据库
	if err := database.DB.Table("user").Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户注册失败"})
		return
	}

	// 模拟生成用户 ID 和 Token
	token := generateJWTToken(user.ID)

	response := UserRegisterResponse{
		StatusCode: 0, // 成功状态码
		StatusMsg:  "注册成功",
		UserID:     user.ID,
		Token:      token,
	}

	c.JSON(http.StatusOK, response)
}

// 加密用户密码
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// 生成 JWT Token
func generateJWTToken(userID int64) string {
	// 创建一个新的Token对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token 过期时间为一天
	})

	// 使用密钥对 Token 进行签名
	tokenString, err := token.SignedString([]byte(config.AppConfigInstance.JWTSecretKey))
	if err != nil {
		// 处理错误
		return ""
	}

	return tokenString
}

func UserLoginHandler(c *gin.Context) {
	// 用于获取用户id
	fmt.Println(c.Get("user_id"))
}
