package config

import (
	"blog-backend/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:ywz0207.@tcp(mysql:3306)/blog?charset=utf8mb4&parseTime=True"
	var err error

	for i := 0; i < 30; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("等待数据库... (%d/30)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	DB.AutoMigrate(&models.Article{}, &models.Category{}, &models.Tag{}, &models.User{})
	log.Println("数据库就绪")
}
