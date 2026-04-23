package controllers

import (
	"blog-backend/config"
	"blog-backend/models"
	"blog-backend/services"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeControllerCSDNSyncProvider struct {
	startResult    *services.CSDNSyncLoginStartResult
	startErr       error
	statusResult   *services.CSDNSyncSession
	statusErr      error
	articlesResult []services.CSDNSyncRemoteArticle
	articlesErr    error
	articleResult  *services.CSDNArticle
	articleErr     error
}

func (p *fakeControllerCSDNSyncProvider) StartLogin() (*services.CSDNSyncLoginStartResult, error) {
	if p.startErr != nil {
		return nil, p.startErr
	}
	if p.startResult == nil {
		return &services.CSDNSyncLoginStartResult{
			Provider:        "csdn",
			ProviderMode:    "fake",
			ProviderSession: "provider-session",
			QRCodeDataURL:   "data:image/png;base64,fake",
			Message:         "ready",
		}, nil
	}
	return p.startResult, nil
}

func (p *fakeControllerCSDNSyncProvider) GetLoginStatus(providerSession string) (*services.CSDNSyncSession, error) {
	if p.statusErr != nil {
		return nil, p.statusErr
	}
	if p.statusResult == nil {
		return &services.CSDNSyncSession{Status: services.CSDNSyncSessionStatusPending, Message: "pending"}, nil
	}
	copied := *p.statusResult
	if copied.Articles != nil {
		copied.Articles = append([]services.CSDNSyncRemoteArticle(nil), copied.Articles...)
	}
	return &copied, nil
}

func (p *fakeControllerCSDNSyncProvider) ListArticles(providerSession string) ([]services.CSDNSyncRemoteArticle, error) {
	if p.articlesErr != nil {
		return nil, p.articlesErr
	}
	return append([]services.CSDNSyncRemoteArticle(nil), p.articlesResult...), nil
}

func (p *fakeControllerCSDNSyncProvider) FetchArticleContent(providerSession string, articleID string) (*services.CSDNArticle, error) {
	if p.articleErr != nil {
		return nil, p.articleErr
	}
	if p.articleResult == nil {
		return nil, errors.New("no article configured")
	}
	copied := *p.articleResult
	return &copied, nil
}

func newAuthenticatedTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Next()
	})
	return r
}

func setupCSDNControllerTestDB(t *testing.T) {
	t.Helper()
	dsn := fmt.Sprintf("file:csdn_controller_%d?mode=memory&cache=private", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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

func TestStartCSDNSyncLoginReturnsSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := services.NewCSDNSyncService(services.NewMemoryCSDNSyncSessionStore(), &fakeControllerCSDNSyncProvider{})
	serviceNow := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	serviceRestore := SetCSDNSyncServiceForTest(service)
	defer serviceRestore()
	_ = serviceNow

	r := newAuthenticatedTestRouter()
	r.POST("/sync/login", StartCSDNSyncLogin)

	req := httptest.NewRequest(http.MethodPost, "/sync/login", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Data services.CSDNSyncSession `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Data.UserID != 7 || resp.Data.ProviderSession != "provider-session" {
		t.Fatalf("unexpected session payload: %+v", resp.Data)
	}
	if resp.Data.QRCodeDataURL == "" {
		t.Fatalf("expected qr code data url in response")
	}
}

func TestGetCSDNSyncSessionReturnsAuthorizedArticles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := services.NewMemoryCSDNSyncSessionStore()
	provider := &fakeControllerCSDNSyncProvider{
		statusResult: &services.CSDNSyncSession{Status: services.CSDNSyncSessionStatusAuthorized, Message: "authorized"},
		articlesResult: []services.CSDNSyncRemoteArticle{
			{ID: "old", Title: "Old", PublishedAt: time.Date(2026, 4, 20, 8, 0, 0, 0, time.UTC)},
			{ID: "new", Title: "New", PublishedAt: time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)},
		},
	}
	service := services.NewCSDNSyncService(store, provider)
	restore := SetCSDNSyncServiceForTest(service)
	defer restore()

	if err := store.Create(&services.CSDNSyncSession{
		ID:              "session-1",
		UserID:          7,
		Provider:        "csdn",
		ProviderMode:    "fake",
		ProviderSession: "provider-session",
		Status:          services.CSDNSyncSessionStatusPending,
		ExpiresAt:       time.Now().Add(time.Minute),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}); err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}

	r := newAuthenticatedTestRouter()
	r.GET("/sync/sessions/:sessionID", GetCSDNSyncSession)

	req := httptest.NewRequest(http.MethodGet, "/sync/sessions/session-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Data services.CSDNSyncSession `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Data.Status != services.CSDNSyncSessionStatusAuthorized {
		t.Fatalf("expected authorized status, got %+v", resp.Data)
	}
	if len(resp.Data.Articles) != 2 || resp.Data.Articles[0].ID != "new" {
		t.Fatalf("expected sorted articles, got %+v", resp.Data.Articles)
	}
}

func TestImportCSDNSyncArticleCreatesDraftArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCSDNControllerTestDB(t)

	if err := config.DB.Create(&models.Category{Name: "Golang"}).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	store := services.NewMemoryCSDNSyncSessionStore()
	provider := &fakeControllerCSDNSyncProvider{
		articleResult: &services.CSDNArticle{
			Title:          "同步导入文章",
			Summary:        "摘要",
			Content:        "## 内容",
			CoverImage:     "https://img.example.com/cover.png",
			Tags:           "Go,CSDN",
			SourceURL:      "https://blog.csdn.net/demo/article/details/1",
			SourcePlatform: "csdn",
		},
	}
	service := services.NewCSDNSyncService(store, provider)
	restore := SetCSDNSyncServiceForTest(service)
	defer restore()

	if err := store.Create(&services.CSDNSyncSession{
		ID:              "session-import",
		UserID:          7,
		Provider:        "csdn",
		ProviderMode:    "fake",
		ProviderSession: "provider-session",
		Status:          services.CSDNSyncSessionStatusAuthorized,
		ExpiresAt:       time.Now().Add(time.Minute),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}); err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}

	r := newAuthenticatedTestRouter()
	r.POST("/sync/import", ImportCSDNSyncArticle)

	body := bytes.NewBufferString(`{"session_id":"session-import","article_id":"remote-1","category_id":1,"status":"draft"}`)
	req := httptest.NewRequest(http.MethodPost, "/sync/import", body)
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
	if article.Title != "同步导入文章" || article.Status != "draft" {
		t.Fatalf("unexpected imported article: %+v", article)
	}
}
