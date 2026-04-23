package services

import (
	"blog-backend/config"
	"blog-backend/models"
	"errors"
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
	service.now = func() time.Time { return time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC) }

	if err := store.Create(&CSDNSyncSession{
		ID: "session-1", UserID: 7, Provider: "csdn", ProviderMode: "fake", ProviderSession: "provider-session",
		Status: CSDNSyncSessionStatusPending, ExpiresAt: time.Date(2026, 4, 23, 10, 2, 0, 0, time.UTC),
		CreatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
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
