package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"blog-backend/services"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCSDNControllerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&models.Article{}, &models.Category{}, &models.Tag{}, &models.User{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}
	config.DB = db
}

func TestPreviewCSDNArticleReturnsParsedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalFetcher := fetchCSDNArticle
	fetchCSDNArticle = func(url string) (*services.CSDNArticle, error) {
		return &services.CSDNArticle{
			Title:          "Go 并发实战",
			Summary:        "讲清 goroutine",
			Content:        "## Go 并发实战",
			CoverImage:     "https://img.example.com/cover.png",
			Tags:           "Go,并发",
			SourceURL:      url,
			SourcePlatform: "csdn",
		}, nil
	}
	defer func() { fetchCSDNArticle = originalFetcher }()

	r := gin.New()
	r.POST("/preview", PreviewCSDNArticle)

	body := bytes.NewBufferString(`{"url":"https://blog.csdn.net/test/article/details/123"}`)
	req := httptest.NewRequest(http.MethodPost, "/preview", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Data services.CSDNArticle `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Data.Title != "Go 并发实战" {
		t.Fatalf("expected title in response, got %+v", resp.Data)
	}
	if resp.Data.SourcePlatform != "csdn" {
		t.Fatalf("expected source platform csdn, got %+v", resp.Data)
	}
}

func TestImportCSDNArticleCreatesDraftArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCSDNControllerTestDB(t)

	if err := config.DB.Create(&models.Category{Name: "Golang"}).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	originalFetcher := fetchCSDNArticle
	fetchCSDNArticle = func(url string) (*services.CSDNArticle, error) {
		return &services.CSDNArticle{
			Title:          "Go 并发实战",
			Summary:        "讲清 goroutine",
			Content:        "## Go 并发实战",
			CoverImage:     "https://img.example.com/cover.png",
			Tags:           "Go,并发",
			SourceURL:      url,
			SourcePlatform: "csdn",
		}, nil
	}
	defer func() { fetchCSDNArticle = originalFetcher }()

	r := gin.New()
	r.POST("/import", ImportCSDNArticle)

	body := bytes.NewBufferString(`{"url":"https://blog.csdn.net/test/article/details/123","category_id":1,"status":"draft"}`)
	req := httptest.NewRequest(http.MethodPost, "/import", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var article models.Article
	if err := config.DB.Preload("Category").First(&article, 1).Error; err != nil {
		t.Fatalf("expected article to be created: %v", err)
	}
	if article.Title != "Go 并发实战" {
		t.Fatalf("expected imported title, got %+v", article)
	}
	if article.Status != "draft" {
		t.Fatalf("expected draft status, got %+v", article)
	}
	if article.Category == nil || article.Category.Name != "Golang" {
		t.Fatalf("expected preloaded category, got %+v", article.Category)
	}
}
