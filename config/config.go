package config

import (
	"github.com/spf13/viper"
)

// AppConfig stores the application configuration
type AppConfig struct {
	DatabaseName             string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	JWTSecretKey             string
	SrcSavedPath             string
	AliyunOSSAccessKeyID     string // 阿里云 OSS Access Key ID
	AliyunOSSAccessKeySecret string // 阿里云 OSS Access Key Secret
	AliyunOSSEndpoint        string // 阿里云 OSS Endpoint
	AliyunOSSBucketName      string // 阿里云 OSS Bucket 名称
}

// AppConfigInstance holds the instance of the application configuration
var AppConfigInstance AppConfig

func LoadConfig() {
	// 初始化 Viper
	viper.SetConfigFile("./config/config.json") // 配置文件的路径

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// 将配置绑定到结构体
	if err := viper.Unmarshal(&AppConfigInstance); err != nil {
		panic(err)
	}
}
