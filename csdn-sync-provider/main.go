package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
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

	log.Printf("csdn-sync-provider skeleton listening on %s", addr)
	if err := http.ListenAndServe(addr, logRequests(mux)); err != nil {
		log.Fatal(err)
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
	writeJSON(w, http.StatusOK, loginStartResponse{
		Provider:        providerName(),
		ProviderMode:    providerMode(),
		ProviderSession: sessionID(),
		QRCodeURL:       qrCodeURL(),
		Message:         envOrDefault("CSDN_PROVIDER_LOGIN_MESSAGE", "provider skeleton ready: 请接入真实 CSDN 登录逻辑"),
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
	writeJSON(w, http.StatusOK, loginStatusResponse{
		Provider:        providerName(),
		ProviderMode:    providerMode(),
		ProviderSession: session,
		Status:          envOrDefault("CSDN_PROVIDER_DEFAULT_STATUS", "pending"),
		Message:         envOrDefault("CSDN_PROVIDER_STATUS_MESSAGE", "waiting for real provider integration"),
		QRCodeURL:       qrCodeURL(),
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
	writeJSON(w, http.StatusOK, articlesResponse{
		Articles: []articleSummary{
			{
				ID:          envOrDefault("CSDN_PROVIDER_ARTICLE_ID", "provider-skeleton-article"),
				Title:       envOrDefault("CSDN_PROVIDER_ARTICLE_TITLE", "CSDN Provider Skeleton Article"),
				Summary:     envOrDefault("CSDN_PROVIDER_ARTICLE_SUMMARY", "独立 provider 服务骨架占位文章，用于 backend 真接口联调。"),
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
	writeJSON(w, http.StatusOK, articleDetailResponse{
		Title:          envOrDefault("CSDN_PROVIDER_ARTICLE_TITLE", "CSDN Provider Skeleton Article"),
		Summary:        envOrDefault("CSDN_PROVIDER_ARTICLE_SUMMARY", "独立 provider 服务骨架占位文章，用于 backend 真接口联调。"),
		Content:        envOrDefault("CSDN_PROVIDER_ARTICLE_CONTENT", "## CSDN Provider Skeleton\n\n当前返回的是独立 provider 服务骨架占位内容，下一步请在这里接入真实扫码、cookie 与文章抓取逻辑。\n\n- article_id: "+articleID+"\n- session: "+session),
		CoverImage:     strings.TrimSpace(os.Getenv("CSDN_PROVIDER_ARTICLE_COVER")),
		Tags:           envOrDefault("CSDN_PROVIDER_ARTICLE_TAGS", "CSDN,Provider,Skeleton"),
		SourceURL:      envOrDefault("CSDN_PROVIDER_ARTICLE_URL", "https://blog.csdn.net/demo/article/details/provider-skeleton"),
		SourcePlatform: providerName(),
	})
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

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
