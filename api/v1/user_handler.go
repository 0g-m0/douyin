// api/v1/auth_handler.go
package v1

import (
	"crypto/rand"
	"douyin/cache"
	"douyin/config"
	"douyin/database"
	"douyin/database/models"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// UserRegisterRequest 是用户注册请求的结构体
type UserRegisterRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// UserLoginRequest 是用户登录请求的结构体
type UserLoginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// UserProfileRequest 是获取用户信息请求的结构体
type UserProfileRequest struct {
	UserID int64  `form:"user_id" binding:"required"`
	Token  string `form:"token" binding:"required"`
}

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
type UserProfileResponse struct {
	StatusCode int32          `json:"status_code"`
	StatusMsg  string         `json:"status_msg"`
	UserDTO    models.UserDTO `json:"user"`
}

// 在全局范围内定义一个互斥锁
var userMutex sync.Mutex

// UserRegisterHandler 处理用户注册请求
func UserRegisterHandler(c *gin.Context) {
	// 获取请求参数
	var request UserRegisterRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	hashPassword, err := hashPassword(request.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户数据模型
	user := models.User{
		Name:     request.Username,
		Password: hashPassword,
	}

	// 使用互斥锁确保只有一个线程可以访问关键代码
	userMutex.Lock()
	defer userMutex.Unlock()

	// 验证用户名是否已经存在
	if err := database.DB.Table("user").Where("name = ?", user.Name).First(&user).Error; err == nil {
		fmt.Println()
		c.JSON(http.StatusOK, UserLoginResponse{
			StatusCode: 1,
			StatusMsg:  "用户名已经存在",
		})
		return
	}

	// 设置用户默认头像
	n, _ := rand.Int(rand.Reader, big.NewInt(9))
	defaultAvatarURL := fmt.Sprintf("https://%s.%s/%s%d%s", config.AppConfigInstance.AliyunOSSBucketName, config.AppConfigInstance.AliyunOSSEndpoint, "img/default_avatar_", n.Int64()+1, ".png")
	user.Avatar = defaultAvatarURL

	// 设置用户默认背景图
	defaultBackgroundImageURL := fmt.Sprintf("https://%s.%s/%s", config.AppConfigInstance.AliyunOSSBucketName, config.AppConfigInstance.AliyunOSSEndpoint, "img/default_background.jpg")
	user.BackgroundImage = defaultBackgroundImageURL

	// 设置用户默认签名
	user.Signature = "这个人很懒，什么都没有留下"

	// 保存用户数据到数据库
	if err := database.DB.Table("user").Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户注册失败"})
		return
	}

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
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token 过期时间为 1 天
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
	// 获取请求参数
	var request UserLoginRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 进行用户登录验证，比对用户名和密码是否正确
	user, err := getUserByUsername(request.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码不正确"})
		return
	}

	// 验证密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
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
	// 获取请求参数
	var request UserProfileRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证user_id与token中的user_id是否一致
	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(int64)
	if userID != request.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID与Token中的用户ID不一致"})
		return
	}

	// 查询用户信息
	var userDTO models.UserDTO
	if err := database.DB.Table("user").Where("id = ?", request.UserID).First(&userDTO).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	userDTO.TotalFavorited, _ = cache.GetTotalFavoritedFromRedis(request.UserID)
	userDTO.FavoriteCount, _ = cache.GetFavoriteCountFromRedis(request.UserID)

	// 构建响应
	response := UserProfileResponse{
		StatusCode: 0, // 成功状态码
		StatusMsg:  "获取用户信息成功",
		UserDTO:    userDTO,
	}

	c.JSON(http.StatusOK, response)
}
