package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

var allowedArticleStatuses = map[string]struct{}{
	"draft":     {},
	"published": {},
}

type articleRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	CoverImage string `json:"cover_image"`
	CategoryID uint   `json:"category_id"`
	Tags       string `json:"tags"`
	IsTop      bool   `json:"is_top"`
	Status     string `json:"status"`
}

func normalizeArticleStatus(status string) (string, bool) {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return "", true
	}
	_, ok := allowedArticleStatuses[status]
	return status, ok
}

func GetArticles(c *gin.Context) {
	var articles []models.Article

	filteredQuery := config.DB.Model(&models.Article{})
	status, ok := normalizeArticleStatus(c.DefaultQuery("status", "published"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章状态，仅支持 draft 或 published"})
		return
	}
	if status != "" {
		filteredQuery = filteredQuery.Where("status = ?", status)
	}

	// 搜索
	if search := c.Query("search"); search != "" {
		filteredQuery = filteredQuery.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 分类筛选
	if catID := c.Query("category_id"); catID != "" {
		filteredQuery = filteredQuery.Where("category_id = ?", catID)
	}

	// 标签筛选
	if tag := strings.TrimSpace(c.Query("tag")); tag != "" {
		filteredQuery = filteredQuery.Where("tags LIKE ?", "%"+tag+"%")
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var total int64
	countQuery := filteredQuery.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计文章数量失败"})
		return
	}

	dataQuery := filteredQuery.Session(&gorm.Session{}).
		Preload("Category").
		Order("is_top desc, created_at desc").
		Limit(pageSize).
		Offset(offset)
	if err := dataQuery.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文章列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetArticle(c *gin.Context) {
	var article models.Article
	if err := config.DB.Preload("Category").First(&article, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	config.DB.Model(&article).Update("views", article.Views+1)
	c.JSON(http.StatusOK, gin.H{"data": article})
}

func CreateArticle(c *gin.Context) {
	var req articleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	article := models.Article{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		CategoryID: req.CategoryID,
		Tags:       req.Tags,
		IsTop:      req.IsTop,
		Status:     status,
	}
	config.DB.Create(&article)
	config.DB.Preload("Category").First(&article, article.ID)
	c.JSON(http.StatusOK, gin.H{"data": article})
}

func UpdateArticle(c *gin.Context) {
	var article models.Article
	if err := config.DB.First(&article, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	var req articleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status, ok := normalizeArticleStatus(req.Status)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章状态，仅支持 draft 或 published"})
		return
	}
	if status == "" {
		status = article.Status
	}
	article.Title = req.Title
	article.Content = req.Content
	article.Summary = req.Summary
	article.CoverImage = req.CoverImage
	article.CategoryID = req.CategoryID
	article.Tags = req.Tags
	article.IsTop = req.IsTop
	article.Status = status
	config.DB.Save(&article)
	config.DB.Preload("Category").First(&article, article.ID)
	c.JSON(http.StatusOK, gin.H{"data": article})
}

func DeleteArticle(c *gin.Context) {
	if err := config.DB.Delete(&models.Article{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
