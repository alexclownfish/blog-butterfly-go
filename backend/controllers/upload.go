package controllers

import (
	"blog-backend/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "上传失败"})
		return
	}
	defer file.Close()

	url, err := utils.UploadToQiniu(file, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传到七牛云失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func ListImages(c *gin.Context) {
	images, err := utils.ListQiniuImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": images})
}

func DeleteImage(c *gin.Context) {
	key := c.Param("key")
	err := utils.DeleteQiniuImage(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
