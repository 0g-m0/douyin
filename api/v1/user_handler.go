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

// UserRegisterResponse 是用户注册响应的结构体
type UserRegisterResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	UserID     int64  `json:"user_id"`
	Token      string `json:"token"`
}

// UserLoginResponse 是用户登录响应的结构体
type UserLoginResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	UserID     int64  `json:"user_id"`
	Token      string `json:"token"`
}

// GetUserProfileResponse 是获取用户信息响应的结构体
type GetUserProfileResponse struct {
	StatusCode int32       `json:"status_code"`
	StatusMsg  string      `json:"status_msg"`
	User       models.User `json:"user"`
}

// UserRegisterHandler 处理用户注册请求
func UserRegisterHandler(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	fmt.Println(username)
	fmt.Println(password)

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
	fmt.Println(1)

	// 验证用户名是否已经存在
	if err := database.DB.Table("user").Where("name = ?", user.Username).First(&user).Error; err == nil {
		fmt.Println()
		c.JSON(http.StatusOK, UserLoginResponse{
			StatusCode: 1,
			StatusMsg:  "用户名已经存在",
		})
		return
	}

	fmt.Println(2)

	// 保存用户数据到数据库
	if err := database.DB.Table("user").Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户注册失败"})
		return
	}

	fmt.Println(3)

	// 生成Token
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

// UserLoginHandler 处理用户登录请求
func UserLoginHandler(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	// 进行用户登录验证，比对用户名和密码是否正确
	user, err := getUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码不正确"})
		return
	}

	// 验证密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码不正确"})
		return
	}

	// 生成 JWT Token
	token := generateJWTToken(user.ID)

	response := UserLoginResponse{
		StatusCode: 0,
		StatusMsg:  "登录成功",
		UserID:     user.ID,
		Token:      token,
	}

	c.JSON(http.StatusOK, response)
}

// 根据用户名查询用户信息
func getUserByUsername(username string) (models.User, error) {
	var user models.User
	if err := database.DB.Table("user").Where("name = ?", username).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// UserInfoHandler 处理用户信息请求
func UserInfoHandler(c *gin.Context) {
	userID := c.Query("user_id")
	token := c.Query("token")

	fmt.Println(userID)

	// 验证token是否有效
	if err := validateToken(userID, token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
		return
	}

	// 查询用户信息
	var user models.User
	if err := database.DB.Table("user").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	// 构建响应
	response := GetUserProfileResponse{
		StatusCode: 0, // 成功状态码
		StatusMsg:  "获取用户信息成功",
		User:       user,
	}

	c.JSON(http.StatusOK, response)
}

func validateToken(userID string, token string) error {
	// 验证Token是否有效
	tokenClaims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfigInstance.JWTSecretKey), nil
	})
	if err != nil || !tokenClaims.Valid {
		return err
	}

	// 验证Token中的用户ID是否与请求的用户ID一致
	claims, _ := tokenClaims.Claims.(jwt.MapClaims)
	if claims["user_id"] != userID {
		return err
	}

	return nil
}
