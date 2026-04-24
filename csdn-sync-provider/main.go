package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"sync"
	"time"
)

type loginStartResponse struct {
	Provider        string `json:"provider"`
	ProviderMode    string `json:"provider_mode"`
	ProviderSession string `json:"provider_session"`
	QRCodeURL       string `json:"qr_code_url"`
	Message         string `json:"message"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type loginStatusResponse struct {
	Provider        string `json:"provider"`
	ProviderMode    string `json:"provider_mode"`
	ProviderSession string `json:"provider_session"`
	Status          string `json:"status"`
	Message         string `json:"message"`
	ErrorMessage    string `json:"error_message,omitempty"`
	QRCodeURL       string `json:"qr_code_url,omitempty"`
}

type articleSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary,omitempty"`
	CoverImage  string `json:"cover_image,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
}

type articlesResponse struct {
	Articles []articleSummary `json:"articles"`
}

type articleDetailResponse struct {
	Title          string `json:"title"`
	Summary        string `json:"summary,omitempty"`
	Content        string `json:"content"`
	CoverImage     string `json:"cover_image,omitempty"`
	Tags           string `json:"tags,omitempty"`
	SourceURL      string `json:"source_url,omitempty"`
	SourcePlatform string `json:"source_platform,omitempty"`
}

type csdnCreateQRResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Data    struct {
		QRCodeURL string `json:"qrCodeUrl"`
		SceneID   string `json:"sceneId"`
	} `json:"data"`
}

type csdnScanStatusResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Data    struct {
		RedirectURL string `json:"redirectUrl"`
	} `json:"data"`
}

type providerSessionState struct {
	ID           string
	SceneID      string
	QRCodeURL    string
	Status       string
	Message      string
	ErrorMessage string
	AuthorizedAt time.Time
	CreatedAt    time.Time
	LastSeenAt   time.Time
	Cookies      []*http.Cookie
}

type providerStore struct {
	mu       sync.RWMutex
	sessions map[string]*providerSessionState
}

type providerService struct {
	client *http.Client
	store  *providerStore
	now    func() time.Time
}

var defaultProviderService = newProviderService()

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/login/start", handleLoginStart)
	mux.HandleFunc("/login/status", handleLoginStatus)
	mux.HandleFunc("/articles", handleArticles)
	mux.HandleFunc("/articles/", handleArticleDetail)

	addr := ":8091"
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		if strings.HasPrefix(port, ":") {
			addr = port
		} else {
			addr = ":" + port
		}
	}

	log.Printf("csdn-sync-provider listening on %s (mode=%s)", addr, providerMode())
	if err := http.ListenAndServe(addr, logRequests(mux)); err != nil {
		log.Fatal(err)
	}
}

func newProviderService() *providerService {
	jar, _ := cookiejar.New(nil)
	return &providerService{
		client: &http.Client{
			Timeout: 20 * time.Second,
			Jar:     jar,
		},
		store: &providerStore{sessions: make(map[string]*providerSessionState)},
		now:   time.Now,
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ok",
		"service":       "csdn-sync-provider",
		"provider_mode": providerMode(),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	})
}

func handleLoginStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if providerMode() != "real" {
		writeJSON(w, http.StatusOK, loginStartResponse{
			Provider:        providerName(),
			ProviderMode:    providerMode(),
			ProviderSession: sessionID(),
			QRCodeURL:       qrCodeURL(),
			Message:         envOrDefault("CSDN_PROVIDER_LOGIN_MESSAGE", "provider skeleton ready: 请接入真实 CSDN 登录逻辑"),
		})
		return
	}

	state, err := defaultProviderService.startLogin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, loginStartResponse{
			Provider:     providerName(),
			ProviderMode: providerMode(),
			Message:      "failed to initialize CSDN login",
			ErrorMessage: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, loginStartResponse{
		Provider:        providerName(),
		ProviderMode:    providerMode(),
		ProviderSession: state.ID,
		QRCodeURL:       state.QRCodeURL,
		Message:         state.Message,
	})
}

func handleLoginStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	session := strings.TrimSpace(r.URL.Query().Get("session"))
	if session == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error_message": "missing session query parameter"})
		return
	}
	if providerMode() != "real" {
		writeJSON(w, http.StatusOK, loginStatusResponse{
			Provider:        providerName(),
			ProviderMode:    providerMode(),
			ProviderSession: session,
			Status:          envOrDefault("CSDN_PROVIDER_DEFAULT_STATUS", "pending"),
			Message:         envOrDefault("CSDN_PROVIDER_STATUS_MESSAGE", "waiting for real provider integration"),
			QRCodeURL:       qrCodeURL(),
		})
		return
	}

	state, err := defaultProviderService.getLoginStatus(r.Context(), session)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, errSessionNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, loginStatusResponse{
			Provider:        providerName(),
			ProviderMode:    providerMode(),
			ProviderSession: session,
			Status:          "failed",
			Message:         "failed to query login status",
			ErrorMessage:    err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, loginStatusResponse{
		Provider:        providerName(),
		ProviderMode:    providerMode(),
		ProviderSession: state.ID,
		Status:          state.Status,
		Message:         state.Message,
		ErrorMessage:    state.ErrorMessage,
		QRCodeURL:       state.QRCodeURL,
	})
}

func handleArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	session := strings.TrimSpace(r.URL.Query().Get("session"))
	if session == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error_message": "missing session query parameter"})
		return
	}
	if providerMode() == "real" {
		state, ok := defaultProviderService.store.get(session)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error_message": "session not found"})
			return
		}
		if state.Status != "authorized" {
			writeJSON(w, http.StatusConflict, map[string]any{"error_message": "session not authorized yet"})
			return
		}
	}
	writeJSON(w, http.StatusOK, articlesResponse{
		Articles: []articleSummary{
			{
				ID:          envOrDefault("CSDN_PROVIDER_ARTICLE_ID", "provider-skeleton-article"),
				Title:       envOrDefault("CSDN_PROVIDER_ARTICLE_TITLE", "CSDN Provider Skeleton Article"),
				Summary:     envOrDefault("CSDN_PROVIDER_ARTICLE_SUMMARY", "当前 provider 已具备真实扫码授权链路，文章抓取仍使用占位数据，下一步继续打通 cookie 抓取。"),
				CoverImage:  strings.TrimSpace(os.Getenv("CSDN_PROVIDER_ARTICLE_COVER")),
				SourceURL:   envOrDefault("CSDN_PROVIDER_ARTICLE_URL", "https://blog.csdn.net/demo/article/details/provider-skeleton"),
				PublishedAt: envOrDefault("CSDN_PROVIDER_ARTICLE_PUBLISHED_AT", time.Date(2026, 4, 24, 3, 0, 0, 0, time.UTC).Format(time.RFC3339)),
			},
		},
	})
}

func handleArticleDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	session := strings.TrimSpace(r.URL.Query().Get("session"))
	if session == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error_message": "missing session query parameter"})
		return
	}
	articleID := strings.TrimPrefix(r.URL.Path, "/articles/")
	articleID = strings.TrimSpace(articleID)
	if articleID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error_message": "missing article id"})
		return
	}
	if providerMode() == "real" {
		state, ok := defaultProviderService.store.get(session)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error_message": "session not found"})
			return
		}
		if state.Status != "authorized" {
			writeJSON(w, http.StatusConflict, map[string]any{"error_message": "session not authorized yet"})
			return
		}
	}
	writeJSON(w, http.StatusOK, articleDetailResponse{
		Title:          envOrDefault("CSDN_PROVIDER_ARTICLE_TITLE", "CSDN Provider Skeleton Article"),
		Summary:        envOrDefault("CSDN_PROVIDER_ARTICLE_SUMMARY", "当前 provider 已具备真实扫码授权链路，文章抓取仍使用占位数据，下一步继续打通 cookie 抓取。"),
		Content:        envOrDefault("CSDN_PROVIDER_ARTICLE_CONTENT", "## CSDN Provider\n\n当前已接入真实扫码授权与轮询状态机；文章抓取正文暂时仍是占位内容。\n\n- article_id: "+articleID+"\n- session: "+session),
		CoverImage:     strings.TrimSpace(os.Getenv("CSDN_PROVIDER_ARTICLE_COVER")),
		Tags:           envOrDefault("CSDN_PROVIDER_ARTICLE_TAGS", "CSDN,Provider,RealLogin"),
		SourceURL:      envOrDefault("CSDN_PROVIDER_ARTICLE_URL", "https://blog.csdn.net/demo/article/details/provider-skeleton"),
		SourcePlatform: providerName(),
	})
}

var errSessionNotFound = errors.New("provider session not found")

func (s *providerService) startLogin(ctx context.Context) (*providerSessionState, error) {
	payload, err := s.createQRCode(ctx)
	if err != nil {
		return nil, err
	}
	sessionID, err := newSessionID()
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	state := &providerSessionState{
		ID:         sessionID,
		SceneID:    strings.TrimSpace(payload.Data.SceneID),
		QRCodeURL:  strings.TrimSpace(payload.Data.QRCodeURL),
		Status:     "pending",
		Message:    envOrDefault("CSDN_PROVIDER_REAL_PENDING_MESSAGE", "请使用 CSDN App / 微信扫码完成授权"),
		CreatedAt:  now,
		LastSeenAt: now,
	}
	s.store.put(state)
	return state, nil
}

func (s *providerService) getLoginStatus(ctx context.Context, sessionID string) (*providerSessionState, error) {
	state, ok := s.store.get(sessionID)
	if !ok {
		return nil, errSessionNotFound
	}
	if state.Status == "authorized" || state.Status == "expired" || state.Status == "failed" {
		return state, nil
	}

	scanResp, cookies, err := s.checkScan(ctx, state.SceneID)
	if err != nil {
		state.Status = "failed"
		state.Message = "query scan status failed"
		state.ErrorMessage = err.Error()
		s.store.put(state)
		return state, nil
	}
	state.Cookies = mergeCookies(state.Cookies, cookies)
	state.LastSeenAt = s.now().UTC()

	if !scanResp.Status {
		switch strings.TrimSpace(scanResp.Code) {
		case "1071":
			state.Status = "expired"
			state.Message = defaultString(strings.TrimSpace(scanResp.Message), "二维码已失效，请重新扫码")
			state.ErrorMessage = state.Message
		default:
			state.Status = "pending"
			state.Message = defaultString(strings.TrimSpace(scanResp.Message), envOrDefault("CSDN_PROVIDER_REAL_PENDING_MESSAGE", "等待扫码确认"))
			state.ErrorMessage = ""
		}
		s.store.put(state)
		return state, nil
	}

	loginResp, loginCookies, err := s.doLogin(ctx, state.SceneID)
	if err != nil {
		state.Status = "failed"
		state.Message = "finalize login failed"
		state.ErrorMessage = err.Error()
		s.store.put(state)
		return state, nil
	}
	state.Cookies = mergeCookies(state.Cookies, loginCookies)
	state.LastSeenAt = s.now().UTC()

	if isAuthorizedResponse(loginResp) {
		state.Status = "authorized"
		state.Message = defaultString(strings.TrimSpace(loginResp.Message), "CSDN 授权成功")
		state.ErrorMessage = ""
		state.AuthorizedAt = s.now().UTC()
		s.store.put(state)
		return state, nil
	}
	if strings.TrimSpace(loginResp.Code) == "1071" {
		state.Status = "expired"
		state.Message = defaultString(strings.TrimSpace(loginResp.Message), "二维码已失效，请重新扫码")
		state.ErrorMessage = state.Message
		s.store.put(state)
		return state, nil
	}
	state.Status = "pending"
	state.Message = defaultString(strings.TrimSpace(loginResp.Message), "扫码成功，等待授权完成")
	state.ErrorMessage = ""
	s.store.put(state)
	return state, nil
}

func (s *providerService) createQRCode(ctx context.Context) (*csdnCreateQRResponse, error) {
	var out csdnCreateQRResponse
	if _, err := s.doJSON(ctx, http.MethodPost, passportBaseURL()+"/v1/register/pc/wxapplets/createQrCode", nil, &out); err != nil {
		return nil, err
	}
	if !out.Status || strings.TrimSpace(out.Data.SceneID) == "" || strings.TrimSpace(out.Data.QRCodeURL) == "" {
		return nil, fmt.Errorf("create qrcode failed: code=%s message=%s", out.Code, out.Message)
	}
	return &out, nil
}

func (s *providerService) checkScan(ctx context.Context, sceneID string) (*csdnScanStatusResponse, []*http.Cookie, error) {
	var out csdnScanStatusResponse
	cookies, err := s.doJSON(ctx, http.MethodPost, passportBaseURL()+"/v1/register/pc/wxapplets/checkScan", map[string]string{"sceneId": sceneID}, &out)
	return &out, cookies, err
}

func (s *providerService) doLogin(ctx context.Context, sceneID string) (*csdnScanStatusResponse, []*http.Cookie, error) {
	var out csdnScanStatusResponse
	cookies, err := s.doJSON(ctx, http.MethodPost, passportBaseURL()+"/v1/register/pc/wxapplets/doLogin", map[string]string{"sceneId": sceneID}, &out)
	return &out, cookies, err
}

func (s *providerService) doJSON(ctx context.Context, method, targetURL string, body any, out any) ([]*http.Cookie, error) {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, targetURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", envOrDefault("CSDN_PROVIDER_USER_AGENT", "Mozilla/5.0 (X11; Linux arm64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0 Safari/537.36"))
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", passportBaseURL())
	req.Header.Set("Referer", passportBaseURL()+"/login")
	if body != nil {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.Cookies(), fmt.Errorf("remote request failed: status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.Cookies(), err
		}
	}
	return resp.Cookies(), nil
}

func (s *providerStore) put(state *providerSessionState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := *state
	s.sessions[state.ID] = &clone
}

func (s *providerStore) get(id string) (*providerSessionState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	clone := *state
	if len(state.Cookies) > 0 {
		clone.Cookies = append([]*http.Cookie(nil), state.Cookies...)
	}
	return &clone, true
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error_message": "method not allowed"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json failed: %v", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}

func providerName() string {
	return envOrDefault("CSDN_PROVIDER_NAME", "csdn")
}

func providerMode() string {
	return envOrDefault("CSDN_PROVIDER_MODE", "skeleton")
}

func sessionID() string {
	return envOrDefault("CSDN_PROVIDER_SESSION_ID", "provider-skeleton-session")
}

func qrCodeURL() string {
	return envOrDefault("CSDN_PROVIDER_QR_CODE_URL", "https://example.com/csdn-provider-skeleton-qr.png")
}

func passportBaseURL() string {
	return strings.TrimRight(envOrDefault("CSDN_PROVIDER_PASSPORT_BASE_URL", "https://passport.csdn.net"), "/")
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func newSessionID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func isAuthorizedResponse(resp *csdnScanStatusResponse) bool {
	if resp == nil {
		return false
	}
	code := strings.TrimSpace(resp.Code)
	if code == "1077" {
		return true
	}
	if strings.TrimSpace(resp.Data.RedirectURL) != "" {
		return true
	}
	return resp.Status && code == "0"
}

func mergeCookies(existing []*http.Cookie, incoming []*http.Cookie) []*http.Cookie {
	if len(incoming) == 0 {
		return existing
	}
	merged := make(map[string]*http.Cookie, len(existing)+len(incoming))
	for _, c := range existing {
		if c != nil {
			merged[c.Name] = c
		}
	}
	for _, c := range incoming {
		if c != nil {
			merged[c.Name] = c
		}
	}
	result := make([]*http.Cookie, 0, len(merged))
	for _, c := range merged {
		result = append(result, c)
	}
	return result
}
