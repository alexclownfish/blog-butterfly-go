package main

import (
	"blog-backend/config"
	"blog-backend/router"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()
	r := gin.Default()
	r.Static("/uploads", "./uploads")
	router.SetupRoutes(r)
	r.Run(":8080")
}
