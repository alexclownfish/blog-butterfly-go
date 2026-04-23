package router

import (
	"blog-backend/controllers"
	"blog-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// 公开接口
		api.GET("/articles", controllers.GetArticles)
		api.GET("/articles/:id", controllers.GetArticle)
		api.GET("/categories", controllers.GetCategories)
		api.GET("/tags", controllers.GetTags)
		api.POST("/login", controllers.Login)

		// 需要认证的接口
		auth := api.Group("", middleware.AuthMiddleware())
		{
			auth.POST("/change-password", controllers.ChangePassword)
			auth.POST("/articles", controllers.CreateArticle)
			auth.PUT("/articles/:id", controllers.UpdateArticle)
			auth.DELETE("/articles/:id", controllers.DeleteArticle)
			auth.POST("/articles/import/csdn/preview", controllers.PreviewCSDNArticle)
			auth.POST("/articles/import/csdn", controllers.ImportCSDNArticle)
			auth.POST("/csdn/sync/login", controllers.StartCSDNSyncLogin)
			auth.GET("/csdn/sync/sessions/:sessionID", controllers.GetCSDNSyncSession)
			auth.POST("/csdn/sync/import", controllers.ImportCSDNSyncArticle)
			auth.POST("/categories", controllers.CreateCategory)
			auth.PUT("/categories/:id", controllers.UpdateCategory)
			auth.DELETE("/categories/:id", controllers.DeleteCategory)
			auth.GET("/dashboard/stats", controllers.GetDashboardStats)
			auth.POST("/upload", controllers.UploadImage)
			auth.GET("/images", controllers.ListImages)
			auth.DELETE("/images/:key", controllers.DeleteImage)
		}
	}
}
