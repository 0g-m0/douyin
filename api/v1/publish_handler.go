// api/v1/publish_handler.go
package v1

import (
	"douyin/config"
	"douyin/database"
	"douyin/database/models"
	"log"
	"os"
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
	"strconv"
	"net/http"
	"time"
	"mime/multipart"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
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

//处理用户发布视频的请求/publish/action
func UserPublishHandler(c *gin.Context) {
	file, err := c.FormFile("data")
    if err != nil {
        // 处理获取文件失败的情况
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 处理获取文件成功的情况
	userIDValue, _ := c.Get("user_id")
	userId, _ := userIDValue.(int64)
	title := c.PostForm("title")

	filename,err :=saveVideo(file,userId)
	if err != nil {
        // 上传视频失败
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
	log.Printf("id:%v的用户上传了视频 %s\n", userId,filename)

	videoURL := fmt.Sprintf("https://%s.%s/%s", config.AppConfigInstance.AliyunOSSBucketName, config.AppConfigInstance.AliyunOSSEndpoint, "video/" + filename)
	fmt.Println(videoURL)
	video := models.Video{
		AuthorUserID: userId,
		PlayURL: videoURL,
		Title: title,
		CreatedAt: time.Now(),
	}

	// 保存用户数据到数据库
	if err := database.DB.Table("video").Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "视频存入数据库失败"})
		return
	}

	response := VideoUploadResponse{
		StatusCode: 0, // 成功状态码
		StatusMsg:  "上传视频成功",
	}

	c.JSON(http.StatusOK, response)

}

// 保存视频到磁盘
func saveVideo(file *multipart.FileHeader,userid int64) (string,error) {
	src, err := file.Open()
	if err != nil {
		fmt.Println("open file fail")
		return "",err
	}
	defer src.Close()

	
	// app_save_directory,err:=getSavePath()
	// if err != nil {
	// 	fmt.Println("上传文件在服务端的保存路径获取失败，查看保存路径配置并确保拥有该路径读写权限")
	// 	return "",err
	// }
	filename:=getUniqueFilename(file.Filename,userid)

	client, err := oss.New(config.AppConfigInstance.AliyunOSSEndpoint, config.AppConfigInstance.AliyunOSSAccessKeyID, config.AppConfigInstance.AliyunOSSAccessKeySecret)
    if err != nil {
        panic(err)
    }

    // 获取存储桶对象
    bucket, err := client.Bucket(config.AppConfigInstance.AliyunOSSBucketName)
    if err != nil {
        panic(err)
    }


    // 调用上传方法
    err = bucket.PutObject("video/" + filename, src)
    if err != nil {
        panic(err)
    }
	// dst, err := os.Create(app_save_directory + "/" + filename)
	// if err != nil {
	// 	fmt.Println("create file fail")
	// 	return "",err
	// }
	// defer dst.Close()
	// fmt.Println(file.Filename)
	// _, err = io.Copy(dst, src)
	// if err != nil {
	// 	fmt.Println("upload file io copy fail")
	// 	return "",err
	// }
	// fmt.Println(filename)
	return filename , nil
}


// 获取文件后缀名
func getFileExtension(fileName string) string {
	dotIndex := strings.LastIndex(fileName, ".")
	if dotIndex == -1 || dotIndex == len(fileName)-1 {
		return ""
	}
	return fileName[dotIndex:]
}

// 获取文件名
func getFileName(fileName string) string {
	dotIndex := strings.LastIndex(fileName, ".")
	if dotIndex == -1 || dotIndex == len(fileName)-1 {
		return ""
	}
	return fileName[:dotIndex]
}

//生成唯一文件名防止同名文件导致无法保存
func getUniqueFilename(Filename string,userid int64) string{
	
	filename := getFileName(Filename)
	fileExten := getFileExtension(Filename)
	timestamp:=time.Now().UnixNano()
	timestamp_str := strconv.FormatInt(timestamp, 10)
	filename += strconv.FormatInt(userid, 10)
	filename += timestamp_str
	filename += fileExten

	return filename
	
}

//获取视频/封面的保存路径
func getSavePath()(string, error) {

	var app_save_directory string
	if config.AppConfigInstance.SrcSavedPath == ""{
		//没有指定特定保存文件夹，默认在主目录
		
		// 获取当前用户的主目录路径
		
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("无法获取主目录路径：", err)
			return 	"",err
		}

		app_save_directory = homeDir + "/mini-douyin-src/"


	}else{
		app_save_directory = config.AppConfigInstance.SrcSavedPath
	}

	

	// 检查文件夹是否存在
	if _, err := os.Stat(app_save_directory); os.IsNotExist(err) {
		// 文件夹不存在，创建文件夹
		err := os.Mkdir(app_save_directory, 0755)
		if err != nil {
			fmt.Println("创建文件夹失败:", err)
			return "",err
		}
		
	} 
		return app_save_directory,nil
	
}