package services

import (
	"blog-backend/config"
	"blog-backend/models"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeCSDNSyncProvider struct {
	startResult    *CSDNSyncLoginStartResult
	startErr       error
	statusResult   *CSDNSyncSession
	statusErr      error
	articlesResult []CSDNSyncRemoteArticle
	articlesErr    error
	articleResult  *CSDNArticle
	articleErr     error
}

func (p *fakeCSDNSyncProvider) StartLogin() (*CSDNSyncLoginStartResult, error) {
	if p.startErr != nil {
		return nil, p.startErr
	}
	if p.startResult == nil {
		return &CSDNSyncLoginStartResult{Provider: "csdn", ProviderMode: "fake", ProviderSession: "provider-session", QRCodeDataURL: "data:image/png;base64,fake"}, nil
	}
	return p.startResult, nil
}

func (p *fakeCSDNSyncProvider) GetLoginStatus(providerSession string) (*CSDNSyncSession, error) {
	if p.statusErr != nil {
		return nil, p.statusErr
	}
	if p.statusResult == nil {
		return &CSDNSyncSession{Status: CSDNSyncSessionStatusPending, Message: "pending"}, nil
	}
	copied := *p.statusResult
	if copied.Articles != nil {
		copied.Articles = append([]CSDNSyncRemoteArticle(nil), copied.Articles...)
	}
	return &copied, nil
}

func (p *fakeCSDNSyncProvider) ListArticles(providerSession string) ([]CSDNSyncRemoteArticle, error) {
	if p.articlesErr != nil {
		return nil, p.articlesErr
	}
	return append([]CSDNSyncRemoteArticle(nil), p.articlesResult...), nil
}

func (p *fakeCSDNSyncProvider) FetchArticleContent(providerSession string, articleID string) (*CSDNArticle, error) {
	if p.articleErr != nil {
		return nil, p.articleErr
	}
	if p.articleResult == nil {
		return nil, errors.New("no article configured")
	}
	copied := *p.articleResult
	return &copied, nil
}

func setupCSDNSyncServiceDB(t *testing.T) {
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

func TestCSDNSyncServiceStartLoginStoresSession(t *testing.T) {
	store := NewMemoryCSDNSyncSessionStore()
	provider := &fakeCSDNSyncProvider{}
	service := NewCSDNSyncService(store, provider)
	fixedNow := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }
	store.now = func() time.Time { return fixedNow }
	service.sessionTTL = 90 * time.Second

	session, err := service.StartLogin(7)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session.UserID != 7 {
		t.Fatalf("expected user id 7, got %+v", session)
	}
	if session.ProviderMode != "fake" {
		t.Fatalf("expected provider mode fake, got %+v", session)
	}
	if session.ExpiresAt != fixedNow.Add(90*time.Second) {
		t.Fatalf("expected custom expiry, got %+v", session)
	}

	stored, err := store.Get(session.ID)
	if err != nil {
		t.Fatalf("expected stored session, got %v", err)
	}
	if stored.QRCodeDataURL == "" {
		t.Fatalf("expected qr data url to be persisted")
	}
}

func TestNewCSDNSyncServiceUsesRealProviderWhenConfigured(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":         "csdn",
			"provider_mode":    "real",
			"provider_session": "provider-session-abc",
			"qr_code_url":      "https://img.example.com/real-qr.png",
			"message":          "real provider ready",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	originalBaseURL, hadBaseURL := os.LookupEnv("CSDN_SYNC_BASE_URL")
	originalMode, hadMode := os.LookupEnv("CSDN_SYNC_PROVIDER_MODE")
	defer func() {
		if hadBaseURL {
			_ = os.Setenv("CSDN_SYNC_BASE_URL", originalBaseURL)
		} else {
			_ = os.Unsetenv("CSDN_SYNC_BASE_URL")
		}
		if hadMode {
			_ = os.Setenv("CSDN_SYNC_PROVIDER_MODE", originalMode)
		} else {
			_ = os.Unsetenv("CSDN_SYNC_PROVIDER_MODE")
		}
	}()

	if err := os.Setenv("CSDN_SYNC_BASE_URL", server.URL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("CSDN_SYNC_PROVIDER_MODE", "real"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	service := NewCSDNSyncService(NewMemoryCSDNSyncSessionStore(), nil)
	session, err := service.StartLogin(7)
	if err != nil {
		t.Fatalf("expected real provider to start login, got %v", err)
	}
	if session.ProviderMode != "real" {
		t.Fatalf("expected real provider mode when env configured, got %+v", session)
	}
	if session.ProviderSession != "provider-session-abc" || session.QRCodeDataURL != "https://img.example.com/real-qr.png" {
		t.Fatalf("expected provider session and qr code from real provider, got %+v", session)
	}
}

func TestCSDNRealProviderStartLoginBuildsQRCodeAndPollsAuthorizedStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST /login/start, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":         "csdn",
			"provider_mode":    "real",
			"provider_session": "provider-session-123",
			"qr_code_url":      "https://img.example.com/qr.png",
			"message":          "请使用 CSDN App 扫码",
		})
	})
	mux.HandleFunc("/login/status", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("session"); got != "provider-session-123" {
			t.Fatalf("expected session query parameter, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":         "csdn",
			"provider_mode":    "real",
			"provider_session": "provider-session-123",
			"status":           "authorized",
			"message":          "扫码成功",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	provider := NewRealCSDNSyncProvider(server.URL)
	loginResult, err := provider.StartLogin()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if loginResult.ProviderMode != "real" {
		t.Fatalf("expected provider mode real, got %+v", loginResult)
	}
	if loginResult.ProviderSession != "provider-session-123" {
		t.Fatalf("expected provider session from api, got %+v", loginResult)
	}
	if loginResult.QRCodeDataURL != "https://img.example.com/qr.png" {
		t.Fatalf("expected qr code url from api, got %+v", loginResult)
	}

	status, err := provider.GetLoginStatus("provider-session-123")
	if err != nil {
		t.Fatalf("expected no error polling status, got %v", err)
	}
	if status.Status != CSDNSyncSessionStatusAuthorized {
		t.Fatalf("expected authorized status, got %+v", status)
	}
	if status.ProviderMode != "real" {
		t.Fatalf("expected real provider mode in status, got %+v", status)
	}
}

func TestCSDNRealProviderListArticlesAndFetchContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("session"); got != "provider-session-123" {
			t.Fatalf("expected session query parameter, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"articles": []map[string]any{
				{
					"id":           "a-1",
					"title":        "Go 并发",
					"summary":      "摘要一",
					"cover_image":  "https://img.example.com/a1.png",
					"source_url":   "https://blog.csdn.net/demo/article/details/1",
					"published_at": "2026-04-22T08:00:00Z",
				},
			},
		})
	})
	mux.HandleFunc("/articles/a-1", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("session"); got != "provider-session-123" {
			t.Fatalf("expected session query parameter, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"title":           "Go 并发",
			"summary":         "摘要一",
			"content":         "## Go 并发",
			"cover_image":     "https://img.example.com/a1.png",
			"tags":            "Go,并发",
			"source_url":      "https://blog.csdn.net/demo/article/details/1",
			"source_platform": "csdn",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	provider := NewRealCSDNSyncProvider(server.URL)
	articles, err := provider.ListArticles("provider-session-123")
	if err != nil {
		t.Fatalf("expected no error listing articles, got %v", err)
	}
	if len(articles) != 1 || articles[0].ID != "a-1" {
		t.Fatalf("expected one remote article, got %+v", articles)
	}
	if articles[0].PublishedAt.IsZero() {
		t.Fatalf("expected published_at to be parsed, got %+v", articles[0])
	}

	article, err := provider.FetchArticleContent("provider-session-123", "a-1")
	if err != nil {
		t.Fatalf("expected no error fetching article, got %v", err)
	}
	if article.Title != "Go 并发" || article.Content != "## Go 并发" {
		t.Fatalf("unexpected article payload: %+v", article)
	}
}

func TestBuildStubQRCodeDataURLReturnsBase64DataURI(t *testing.T) {
	value := buildStubQRCodeDataURL("provider-session")
	const prefix = "data:image/svg+xml;base64,"
	if !strings.HasPrefix(value, prefix) {
		t.Fatalf("expected base64 data uri prefix, got %q", value)
	}
	encoded := strings.TrimPrefix(value, prefix)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("expected valid base64 payload, got %v", err)
	}
	decodedText := string(decoded)
	if !strings.Contains(decodedText, "Stub QR") {
		t.Fatalf("expected svg label in payload, got %q", decodedText)
	}
	if !strings.Contains(decodedText, "provider-session") {
		t.Fatalf("expected provider session in payload, got %q", decodedText)
	}
}

func TestCSDNSyncServiceRefreshSessionLoadsAuthorizedArticles(t *testing.T) {
	store := NewMemoryCSDNSyncSessionStore()
	provider := &fakeCSDNSyncProvider{
		statusResult: &CSDNSyncSession{Status: CSDNSyncSessionStatusAuthorized, Message: "authorized"},
		articlesResult: []CSDNSyncRemoteArticle{
			{ID: "older", Title: "旧文章", PublishedAt: time.Date(2026, 4, 20, 8, 0, 0, 0, time.UTC)},
			{ID: "newer", Title: "新文章", PublishedAt: time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)},
		},
	}
	service := NewCSDNSyncService(store, provider)
	fixedNow := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }
	store.now = func() time.Time { return fixedNow }

	if err := store.Create(&CSDNSyncSession{
		ID: "session-1", UserID: 7, Provider: "csdn", ProviderMode: "fake", ProviderSession: "provider-session",
		Status: CSDNSyncSessionStatusPending, ExpiresAt: time.Date(2026, 4, 23, 10, 2, 0, 0, time.UTC),
		CreatedAt: fixedNow, UpdatedAt: fixedNow,
	}); err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}

	session, err := service.RefreshSession(7, "session-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session.Status != CSDNSyncSessionStatusAuthorized {
		t.Fatalf("expected authorized session, got %+v", session)
	}
	if len(session.Articles) != 2 || session.Articles[0].ID != "newer" {
		t.Fatalf("expected sorted articles, got %+v", session.Articles)
	}
}

func TestCSDNSyncServiceImportArticleCreatesArticle(t *testing.T) {
	setupCSDNSyncServiceDB(t)
	if err := config.DB.Create(&models.Category{Name: "Golang"}).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	store := NewMemoryCSDNSyncSessionStore()
	provider := &fakeCSDNSyncProvider{
		articleResult: &CSDNArticle{
			Title: "同步导入文章", Summary: "摘要", Content: "## 内容", CoverImage: "https://img.example.com/cover.png", Tags: "Go,CSDN", SourceURL: "https://blog.csdn.net/demo/article/details/1", SourcePlatform: "csdn",
		},
	}
	service := NewCSDNSyncService(store, provider)
	if err := store.Create(&CSDNSyncSession{
		ID: "session-import", UserID: 7, Provider: "csdn", ProviderMode: "fake", ProviderSession: "provider-session",
		Status: CSDNSyncSessionStatusAuthorized, ExpiresAt: time.Now().Add(time.Minute), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}

	article, err := service.ImportArticle(7, "session-import", 1, "draft", "remote-article-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if article.ID == 0 || article.Title != "同步导入文章" {
		t.Fatalf("unexpected article result: %+v", article)
	}
	if article.Category == nil || article.Category.Name != "Golang" {
		t.Fatalf("expected preloaded category, got %+v", article.Category)
	}
}

func TestMemoryCSDNSyncSessionStoreExpiresSession(t *testing.T) {
	store := NewMemoryCSDNSyncSessionStore()
	baseNow := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return baseNow }
	if err := store.Create(&CSDNSyncSession{ID: "expired-1", Status: CSDNSyncSessionStatusPending, ExpiresAt: baseNow.Add(-time.Second), UpdatedAt: baseNow.Add(-10 * time.Minute)}); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	store.now = func() time.Time { return baseNow }

	_, err := store.Get("expired-1")
	if !errors.Is(err, ErrCSDNSyncSessionNotFound) {
		t.Fatalf("expected not found after cleanup, got %v", err)
	}
}
