package config

import (
	"blog-backend/models"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "ChangeMe123!"
)

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
	initDefaultAdmin()
	log.Println("数据库就绪")
}

func initDefaultAdmin() {
	username := getenvDefault("DEFAULT_ADMIN_USERNAME", defaultAdminUsername)
	password, passwordFromEnv := getenvWithSource("DEFAULT_ADMIN_PASSWORD", defaultAdminPassword)

	var user models.User
	err := DB.Where("username = ?", username).First(&user).Error
	if err == nil {
		log.Printf("默认管理员已存在，跳过初始化: %s", username)
		return
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("检查默认管理员失败: %v", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("默认管理员密码加密失败: %v", err)
		return
	}

	user = models.User{
		Username: username,
		Password: string(hashedPassword),
	}
	if err := DB.Create(&user).Error; err != nil {
		log.Printf("默认管理员创建失败: %v", err)
		return
	}

	if passwordFromEnv {
		log.Printf("默认管理员初始化完成: username=%s (密码来自环境变量 DEFAULT_ADMIN_PASSWORD)", username)
	} else {
		log.Printf("默认管理员初始化完成: username=%s (当前使用内置默认密码，请尽快登录后修改，或通过 DEFAULT_ADMIN_PASSWORD 覆盖)", username)
	}
}

func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvWithSource(key, fallback string) (string, bool) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, false
	}
	return value, true
}
