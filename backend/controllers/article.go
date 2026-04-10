package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetArticles(c *gin.Context) {
	var articles []models.Article
	query := config.DB.Preload("Category").Order("is_top desc, created_at desc")
	
	// 搜索
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// 分类筛选
	if catID := c.Query("category_id"); catID != "" {
		query = query.Where("category_id = ?", catID)
	}
	
	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize
	
	var total int64
	query.Model(&models.Article{}).Count(&total)
	query.Limit(pageSize).Offset(offset).Find(&articles)
	
	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"total": total,
		"page": page,
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
	var article models.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&article)
	c.JSON(http.StatusOK, gin.H{"data": article})
}

func UpdateArticle(c *gin.Context) {
	var article models.Article
	if err := config.DB.First(&article, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Save(&article)
	c.JSON(http.StatusOK, gin.H{"data": article})
}

func DeleteArticle(c *gin.Context) {
	if err := config.DB.Delete(&models.Article{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
