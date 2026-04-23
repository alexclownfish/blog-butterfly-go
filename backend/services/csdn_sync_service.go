package services

import (
	"blog-backend/config"
	"blog-backend/models"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const DefaultCSDNSyncSessionTTL = 2 * time.Minute

type CSDNSyncProvider interface {
	StartLogin() (*CSDNSyncLoginStartResult, error)
	GetLoginStatus(providerSession string) (*CSDNSyncSession, error)
	ListArticles(providerSession string) ([]CSDNSyncRemoteArticle, error)
	FetchArticleContent(providerSession string, articleID string) (*CSDNArticle, error)
}

type StubCSDNSyncProvider struct{}

func (p *StubCSDNSyncProvider) StartLogin() (*CSDNSyncLoginStartResult, error) {
	providerSession, err := newCSDNSyncSessionID()
	if err != nil {
		return nil, err
	}
	return &CSDNSyncLoginStartResult{
		Provider:        "csdn",
		ProviderMode:    "stub",
		ProviderSession: providerSession,
		QRCodeDataURL:   buildStubQRCodeDataURL(providerSession),
		Message:         "开发占位模式：扫码能力待接入真实 CSDN 登录接口",
	}, nil
}

func (p *StubCSDNSyncProvider) GetLoginStatus(providerSession string) (*CSDNSyncSession, error) {
	return &CSDNSyncSession{
		Provider:        "csdn",
		ProviderMode:    "stub",
		ProviderSession: providerSession,
		Status:          CSDNSyncSessionStatusPending,
		Message:         "等待扫码确认（stub）",
	}, nil
}

func (p *StubCSDNSyncProvider) ListArticles(providerSession string) ([]CSDNSyncRemoteArticle, error) {
	return []CSDNSyncRemoteArticle{
		{
			ID:          "stub-go-concurrency",
			Title:       "Go 并发实战（示例）",
			Summary:     "用于前后端联调的占位文章列表",
			CoverImage:  "https://img.example.com/csdn-sync-cover.png",
			SourceURL:   "https://blog.csdn.net/demo/article/details/10001",
			PublishedAt: time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:          "stub-k3s-monitoring",
			Title:       "K3s 监控排障笔记（示例）",
			Summary:     "第二篇示例文章，验证列表展示与导入动作",
			CoverImage:  "",
			SourceURL:   "https://blog.csdn.net/demo/article/details/10002",
			PublishedAt: time.Date(2026, 4, 20, 12, 30, 0, 0, time.UTC),
		},
	}, nil
}

func (p *StubCSDNSyncProvider) FetchArticleContent(providerSession string, articleID string) (*CSDNArticle, error) {
	articles, err := p.ListArticles(providerSession)
	if err != nil {
		return nil, err
	}
	for _, article := range articles {
		if article.ID != articleID {
			continue
		}
		content := fmt.Sprintf("## %s\n\n> 当前为开发占位内容，用于打通扫码同步导入完整链路。\n\n原文链接：%s\n", article.Title, article.SourceURL)
		return &CSDNArticle{
			Title:          article.Title,
			Summary:        article.Summary,
			Content:        content,
			CoverImage:     article.CoverImage,
			Tags:           "CSDN,同步导入,占位数据",
			SourceURL:      article.SourceURL,
			SourcePlatform: "csdn",
		}, nil
	}
	return nil, errors.New("未找到指定 CSDN 文章")
}

type CSDNSyncService struct {
	store      CSDNSyncSessionStore
	provider   CSDNSyncProvider
	sessionTTL time.Duration
	now        func() time.Time
}

func NewCSDNSyncService(store CSDNSyncSessionStore, provider CSDNSyncProvider) *CSDNSyncService {
	if store == nil {
		store = NewMemoryCSDNSyncSessionStore()
	}
	if provider == nil {
		provider = &StubCSDNSyncProvider{}
	}
	return &CSDNSyncService{
		store:      store,
		provider:   provider,
		sessionTTL: DefaultCSDNSyncSessionTTL,
		now:        time.Now,
	}
}

func (s *CSDNSyncService) StartLogin(userID uint) (*CSDNSyncSession, error) {
	if userID == 0 {
		return nil, errors.New("未找到当前登录用户")
	}
	startResult, err := s.provider.StartLogin()
	if err != nil {
		return nil, err
	}
	now := s.now()
	sessionID, err := newCSDNSyncSessionID()
	if err != nil {
		return nil, err
	}
	session := &CSDNSyncSession{
		ID:              sessionID,
		UserID:          userID,
		Provider:        defaultString(startResult.Provider, "csdn"),
		ProviderMode:    defaultString(startResult.ProviderMode, "stub"),
		ProviderSession: startResult.ProviderSession,
		Status:          CSDNSyncSessionStatusPending,
		Message:         startResult.Message,
		QRCodeDataURL:   startResult.QRCodeDataURL,
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       now.Add(s.sessionTTL),
	}
	if err := s.store.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *CSDNSyncService) GetSession(userID uint, sessionID string) (*CSDNSyncSession, error) {
	session, err := s.store.Get(strings.TrimSpace(sessionID))
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, ErrCSDNSyncSessionNotFound
	}
	return session, nil
}

func (s *CSDNSyncService) RefreshSession(userID uint, sessionID string) (*CSDNSyncSession, error) {
	session, err := s.GetSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	if !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(now) {
		session.Status = CSDNSyncSessionStatusExpired
		session.ErrorMessage = "登录会话已过期，请重新扫码"
		session.UpdatedAt = now
		_ = s.store.Update(session)
		return session, nil
	}

	providerState, err := s.provider.GetLoginStatus(session.ProviderSession)
	if err != nil {
		session.Status = CSDNSyncSessionStatusFailed
		session.ErrorMessage = err.Error()
		session.UpdatedAt = now
		_ = s.store.Update(session)
		return session, err
	}

	if providerState.Status != "" {
		session.Status = providerState.Status
	}
	if providerState.Message != "" {
		session.Message = providerState.Message
	}
	if providerState.ErrorMessage != "" {
		session.ErrorMessage = providerState.ErrorMessage
	}
	if providerState.QRCodeDataURL != "" {
		session.QRCodeDataURL = providerState.QRCodeDataURL
	}
	if len(providerState.Articles) > 0 {
		session.Articles = append([]CSDNSyncRemoteArticle(nil), providerState.Articles...)
	}

	if session.Status == CSDNSyncSessionStatusAuthorized {
		articles, listErr := s.provider.ListArticles(session.ProviderSession)
		if listErr != nil {
			session.Status = CSDNSyncSessionStatusFailed
			session.ErrorMessage = listErr.Error()
		} else {
			sort.SliceStable(articles, func(i, j int) bool {
				return articles[i].PublishedAt.After(articles[j].PublishedAt)
			})
			session.Articles = articles
		}
	}

	session.UpdatedAt = now
	if err := s.store.Update(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *CSDNSyncService) ImportArticle(userID uint, sessionID string, categoryID uint, status string, articleID string) (*models.Article, error) {
	if categoryID == 0 {
		return nil, errors.New("请选择文章分类")
	}
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		status = "draft"
	}
	if status != "draft" && status != "published" {
		return nil, errors.New("无效的文章状态，仅支持 draft 或 published")
	}

	session, err := s.GetSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != CSDNSyncSessionStatusAuthorized {
		return nil, errors.New("当前登录会话尚未授权完成")
	}

	articleData, err := s.provider.FetchArticleContent(session.ProviderSession, strings.TrimSpace(articleID))
	if err != nil {
		return nil, err
	}

	article := models.Article{
		Title:      articleData.Title,
		Content:    articleData.Content,
		Summary:    articleData.Summary,
		CoverImage: articleData.CoverImage,
		CategoryID: categoryID,
		Tags:       articleData.Tags,
		Status:     status,
	}
	if err := config.DB.Create(&article).Error; err != nil {
		return nil, errors.New("导入文章失败，请检查分类是否存在或数据是否有效")
	}
	if err := config.DB.Preload("Category").First(&article, article.ID).Error; err != nil {
		return nil, errors.New("文章已导入，但加载详情失败")
	}
	return &article, nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func newCSDNSyncSessionID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func buildStubQRCodeDataURL(content string) string {
	return "data:image/svg+xml;utf8," + fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="220" height="220" viewBox="0 0 220 220"><rect width="220" height="220" rx="24" fill="#0f172a"/><rect x="24" y="24" width="172" height="172" rx="18" fill="#ffffff"/><text x="110" y="86" text-anchor="middle" font-size="18" fill="#111827" font-family="Arial">CSDN</text><text x="110" y="116" text-anchor="middle" font-size="18" fill="#111827" font-family="Arial">Stub QR</text><text x="110" y="150" text-anchor="middle" font-size="11" fill="#475569" font-family="Arial">%s</text></svg>`,
		content,
	)
}
