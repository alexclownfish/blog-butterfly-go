package middleware

import (
	"blog-backend/config"
	"blog-backend/models"
	"blog-backend/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token无效"})
			c.Abort()
			return
		}

		var user models.User
		if err := config.DB.Select("id", "username", "force_password_change").First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已失效"})
			c.Abort()
			return
		}

		changePasswordPath := "/api/change-password"
		if user.ForcePasswordChange && c.FullPath() != changePasswordPath {
			c.JSON(http.StatusForbidden, gin.H{
				"error":                 "首次登录后必须先修改密码",
				"force_password_change": true,
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", user.Username)
		c.Set("force_password_change", user.ForcePasswordChange)
		c.Next()
	}
}
