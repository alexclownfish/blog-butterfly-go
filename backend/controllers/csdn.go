package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"blog-backend/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var fetchCSDNArticle = services.FetchCSDNArticle

type csdnPreviewRequest struct {
	URL string `json:"url" binding:"required"`
}

type csdnImportRequest struct {
	URL        string `json:"url" binding:"required"`
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
