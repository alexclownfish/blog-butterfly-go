package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"blog-backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type categoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type dashboardStatsResponse struct {
	ArticleTotal   int64 `json:"article_total"`
	PublishedTotal int64 `json:"published_total"`
	DraftTotal     int64 `json:"draft_total"`
	CategoryTotal  int64 `json:"category_total"`
	ImageTotal     int   `json:"image_total"`
	TopArticleTotal int64 `json:"top_article_total"`
}

func GetCategories(c *gin.Context) {
	var categories []models.Category
	config.DB.Order("name asc").Find(&categories)
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func GetTags(c *gin.Context) {
	var tags []models.Tag
	config.DB.Find(&tags)
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func CreateCategory(c *gin.Context) {
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := models.Category{Name: strings.TrimSpace(req.Name)}
	if category.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分类名称不能为空"})
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建分类失败，名称可能已存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": category})
}

func UpdateCategory(c *gin.Context) {
	var category models.Category
	if err := config.DB.First(&category, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		return
	}

	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分类名称不能为空"})
		return
	}

	category.Name = name
	if err := config.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "更新分类失败，名称可能已存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": category})
}

func DeleteCategory(c *gin.Context) {
	if err := config.DB.Delete(&models.Category{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "删除分类失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetDashboardStats(c *gin.Context) {
	stats := dashboardStatsResponse{}

	if err := config.DB.Model(&models.Article{}).Count(&stats.ArticleTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计文章数量失败"})
		return
	}
	if err := config.DB.Model(&models.Article{}).Where("status = ?", "published").Count(&stats.PublishedTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计已发布文章失败"})
		return
	}
	if err := config.DB.Model(&models.Article{}).Where("status = ?", "draft").Count(&stats.DraftTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计草稿数量失败"})
		return
	}
	if err := config.DB.Model(&models.Article{}).Where("is_top = ?", true).Count(&stats.TopArticleTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计置顶文章失败"})
		return
	}
	if err := config.DB.Model(&models.Category{}).Count(&stats.CategoryTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计分类数量失败"})
		return
	}

	images, err := utils.ListQiniuImages()
	if err == nil {
		stats.ImageTotal = len(images)
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
