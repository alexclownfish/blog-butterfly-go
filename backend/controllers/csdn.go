package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"blog-backend/services"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var fetchCSDNArticle = services.FetchCSDNArticle
var csdnSyncService = services.NewCSDNSyncService(nil, nil)

type csdnPreviewRequest struct {
	URL string `json:"url" binding:"required"`
}

type csdnImportRequest struct {
	URL        string `json:"url" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
	Status     string `json:"status"`
}

type csdnSyncImportRequest struct {
	SessionID  string `json:"session_id" binding:"required"`
	ArticleID  string `json:"article_id" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
	Status     string `json:"status"`
}

func PreviewCSDNArticle(c *gin.Context) {
	var req csdnPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	article, err := fetchCSDNArticle(strings.TrimSpace(req.URL))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

func ImportCSDNArticle(c *gin.Context) {
	var req csdnImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	status, ok := normalizeArticleStatus(req.Status)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章状态，仅支持 draft 或 published"})
		return
	}
	if status == "" {
		status = "draft"
	}
	if req.CategoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文章分类"})
		return
	}

	articleData, err := fetchCSDNArticle(strings.TrimSpace(req.URL))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article := models.Article{
		Title:      articleData.Title,
		Content:    articleData.Content,
		Summary:    articleData.Summary,
		CoverImage: articleData.CoverImage,
		CategoryID: req.CategoryID,
		Tags:       articleData.Tags,
		Status:     status,
	}
	if err := config.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "导入文章失败，请检查分类是否存在或数据是否有效"})
		return
	}
	if err := config.DB.Preload("Category").First(&article, article.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文章已导入，但加载详情失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

func StartCSDNSyncLogin(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	session, err := csdnSyncService.StartLogin(userID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": session})
}

func GetCSDNSyncSession(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	sessionID := strings.TrimSpace(c.Param("sessionID"))
	session, err := csdnSyncService.RefreshSession(userID, sessionID)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, services.ErrCSDNSyncSessionNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": session})
}

func ImportCSDNSyncArticle(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req csdnSyncImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	status, ok := normalizeArticleStatus(req.Status)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章状态，仅支持 draft 或 published"})
		return
	}
	if status == "" {
		status = "draft"
	}

	article, err := csdnSyncService.ImportArticle(userID, req.SessionID, req.CategoryID, status, req.ArticleID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCSDNSyncSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case err.Error() == "请选择文章分类" || err.Error() == "当前登录会话尚未授权完成" || err.Error() == "无效的文章状态，仅支持 draft 或 published":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": article})
}

func currentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	uid, ok := userID.(uint)
	return uid, ok
}

func SetCSDNSyncServiceForTest(service *services.CSDNSyncService) func() {
	original := csdnSyncService
	if service == nil {
		csdnSyncService = services.NewCSDNSyncService(nil, nil)
	} else {
		csdnSyncService = service
	}
	return func() {
		csdnSyncService = original
	}
}
