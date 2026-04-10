package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetCategories(c *gin.Context) {
	var categories []models.Category
	config.DB.Find(&categories)
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func GetTags(c *gin.Context) {
	var tags []models.Tag
	config.DB.Find(&tags)
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&category)
	c.JSON(http.StatusOK, gin.H{"data": category})
}

func DeleteCategory(c *gin.Context) {
	config.DB.Delete(&models.Category{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
