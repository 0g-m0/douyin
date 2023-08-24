// database/database.go
package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

func ConnectDB() error {
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"root",
		"root",
		"127.0.0.1",
		"3306",
		"douyin",
		// "remote_user",
		// "remote_password",
		// "8.130.126.94",
		// "3306",
		// "douyin",
	)

	var err error
	// DB, err = gorm.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/douyin?charset=utf8&parseTime=true")

	DB, err = gorm.Open("mysql", dbURL)
	if err != nil {
		panic("Failed to connect to the database")
	} else {
		fmt.Println("Connected to the database")
	}

	return nil
}