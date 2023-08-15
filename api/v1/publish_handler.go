// api/v1/publish_handler.go
package v1

import (
	// "douyin/config"
	// "douyin/database"
	// "douyin/database/models"
	// "fmt"
	// "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	// "golang.org/x/crypto/bcrypt"
	// "net/http"
	// "time"
	"mime/multipart"
)
//上传视频request结构体
type VideoUploadRequest struct {
	Token  string                `form:"token" binding:"required"`
	File   *multipart.FileHeader `form:"file" binding:"required"`
	Title  string                `form:"title" binding:"required,max=50"`
}
//上传视频后服务端response结构体
type VideoUploadResponse struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

func UserPublishHandler(c *gin.Context) {

}