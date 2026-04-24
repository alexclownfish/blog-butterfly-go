package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/login/start", handleLoginStart)
	mux.HandleFunc("/login/status", handleLoginStatus)
	mux.HandleFunc("/articles", handleArticles)
	mux.HandleFunc("/articles/", handleArticleDetail)
	return httptest.NewServer(mux)
}

func TestHealthEndpoint(t *testing.T) {
	defaultProviderService = newProviderService()
	t.Setenv("CSDN_PROVIDER_MODE", "skeleton")
	server := newTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode health response failed: %v", err)
	}
	if payload["service"] != "csdn-sync-provider" {
		t.Fatalf("expected service name, got %+v", payload)
	}
	if payload["provider_mode"] != "skeleton" {
		t.Fatalf("expected provider_mode skeleton, got %+v", payload)
	}
}

func TestLoginStartUsesEnvDrivenPayload(t *testing.T) {
	defaultProviderService = newProviderService()
	t.Setenv("CSDN_PROVIDER_NAME", "csdn")
	t.Setenv("CSDN_PROVIDER_MODE", "skeleton")
	t.Setenv("CSDN_PROVIDER_SESSION_ID", "session-from-env")
	t.Setenv("CSDN_PROVIDER_QR_CODE_URL", "https://example.com/qr.png")
	t.Setenv("CSDN_PROVIDER_LOGIN_MESSAGE", "scan me maybe")
	server := newTestServer()
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/login/start", nil)
	if err != nil {
		t.Fatalf("build request failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login start request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload loginStartResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode login start response failed: %v", err)
	}
	if payload.ProviderSession != "session-from-env" {
		t.Fatalf("expected session from env, got %+v", payload)
	}
	if payload.QRCodeURL != "https://example.com/qr.png" {
		t.Fatalf("expected qr code url from env, got %+v", payload)
	}
	if payload.Message != "scan me maybe" {
		t.Fatalf("expected message from env, got %+v", payload)
	}
}

func TestLoginStatusRequiresSessionQuery(t *testing.T) {
	defaultProviderService = newProviderService()
	server := newTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/login/status")
	if err != nil {
		t.Fatalf("login status request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestArticlesAndDetailEndpointsReturnConfiguredArticle(t *testing.T) {
	defaultProviderService = newProviderService()
	t.Setenv("CSDN_PROVIDER_ARTICLE_ID", "article-42")
	t.Setenv("CSDN_PROVIDER_ARTICLE_TITLE", "Provider Test Article")
	t.Setenv("CSDN_PROVIDER_ARTICLE_SUMMARY", "summary from env")
	t.Setenv("CSDN_PROVIDER_ARTICLE_URL", "https://blog.csdn.net/demo/article/details/42")
	t.Setenv("CSDN_PROVIDER_ARTICLE_TAGS", "Go,Test")
	t.Setenv("CSDN_PROVIDER_ARTICLE_CONTENT", "## Provider Test\n\nHello from test")
	server := newTestServer()
	defer server.Close()

	articlesResp, err := http.Get(server.URL + "/articles?session=test-session")
	if err != nil {
		t.Fatalf("articles request failed: %v", err)
	}
	defer articlesResp.Body.Close()

	if articlesResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /articles, got %d", articlesResp.StatusCode)
	}

	var articlesPayload articlesResponse
	if err := json.NewDecoder(articlesResp.Body).Decode(&articlesPayload); err != nil {
		t.Fatalf("decode articles response failed: %v", err)
	}
	if len(articlesPayload.Articles) != 1 {
		t.Fatalf("expected one article, got %+v", articlesPayload)
	}
	if articlesPayload.Articles[0].ID != "article-42" {
		t.Fatalf("expected configured article id, got %+v", articlesPayload.Articles[0])
	}

	detailResp, err := http.Get(server.URL + "/articles/article-42?session=test-session")
	if err != nil {
		t.Fatalf("article detail request failed: %v", err)
	}
	defer detailResp.Body.Close()

	if detailResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for detail, got %d", detailResp.StatusCode)
	}

	var detailPayload articleDetailResponse
	if err := json.NewDecoder(detailResp.Body).Decode(&detailPayload); err != nil {
		t.Fatalf("decode article detail response failed: %v", err)
	}
	if detailPayload.Title != "Provider Test Article" {
		t.Fatalf("expected configured title, got %+v", detailPayload)
	}
	if !strings.Contains(detailPayload.Content, "Provider Test") {
		t.Fatalf("expected configured content, got %+v", detailPayload)
	}
	if detailPayload.SourcePlatform != "csdn" {
		t.Fatalf("expected source platform csdn, got %+v", detailPayload)
	}
}

func TestRealLoginFlowUsesRemoteQRCodeAndBecomesAuthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/register/pc/wxapplets/createQrCode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST createQrCode, got %s", r.Method)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  true,
			"message": "ok",
			"code":    "0",
			"data": map[string]any{
				"qrCodeUrl": "https://passport.csdn.net/qrcode/test.png",
				"sceneId":   "scene-123",
			},
		})
	})
	mux.HandleFunc("/v1/register/pc/wxapplets/checkScan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST checkScan, got %s", r.Method)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode checkScan body failed: %v", err)
		}
		if payload["sceneId"] != "scene-123" {
			t.Fatalf("expected sceneId scene-123, got %+v", payload)
		}
		http.SetCookie(w, &http.Cookie{Name: "scan_cookie", Value: "scan-ok", Path: "/"})
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  true,
			"message": "scanned",
			"code":    "1077",
			"data": map[string]any{
				"redirectUrl": "https://passport.csdn.net/login/success",
			},
		})
	})
	mux.HandleFunc("/v1/register/pc/wxapplets/doLogin", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode doLogin body failed: %v", err)
		}
		if payload["sceneId"] != "scene-123" {
			t.Fatalf("expected sceneId scene-123, got %+v", payload)
		}
		http.SetCookie(w, &http.Cookie{Name: "UserName", Value: "demo-user", Path: "/"})
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  true,
			"message": "authorized",
			"code":    "1077",
			"data": map[string]any{
				"redirectUrl": "https://blog.csdn.net/demo-user",
			},
		})
	})
	passport := httptest.NewServer(mux)
	defer passport.Close()

	defaultProviderService = newProviderService()
	t.Setenv("CSDN_PROVIDER_MODE", "real")
	t.Setenv("CSDN_PROVIDER_PASSPORT_BASE_URL", passport.URL)
	server := newTestServer()
	defer server.Close()

	startReq, _ := http.NewRequest(http.MethodPost, server.URL+"/login/start", nil)
	startResp, err := http.DefaultClient.Do(startReq)
	if err != nil {
		t.Fatalf("login start request failed: %v", err)
	}
	defer startResp.Body.Close()
	if startResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from login start, got %d", startResp.StatusCode)
	}
	var startPayload loginStartResponse
	if err := json.NewDecoder(startResp.Body).Decode(&startPayload); err != nil {
		t.Fatalf("decode login start failed: %v", err)
	}
	if startPayload.ProviderSession == "" {
		t.Fatalf("expected provider session, got %+v", startPayload)
	}
	if startPayload.QRCodeURL != "https://passport.csdn.net/qrcode/test.png" {
		t.Fatalf("expected real qr url, got %+v", startPayload)
	}

	statusResp, err := http.Get(server.URL + "/login/status?session=" + startPayload.ProviderSession)
	if err != nil {
		t.Fatalf("login status request failed: %v", err)
	}
	defer statusResp.Body.Close()
	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from login status, got %d", statusResp.StatusCode)
	}
	var statusPayload loginStatusResponse
	if err := json.NewDecoder(statusResp.Body).Decode(&statusPayload); err != nil {
		t.Fatalf("decode login status failed: %v", err)
	}
	if statusPayload.Status != "authorized" {
		t.Fatalf("expected authorized, got %+v", statusPayload)
	}
	stored, ok := defaultProviderService.store.get(startPayload.ProviderSession)
	if !ok {
		t.Fatalf("expected stored session %s", startPayload.ProviderSession)
	}
	if stored.SceneID != "scene-123" {
		t.Fatalf("expected stored scene id, got %+v", stored)
	}
	if len(stored.Cookies) == 0 {
		t.Fatalf("expected cookies to be stored, got %+v", stored)
	}
}

func TestRealModeArticlesRequireAuthorizedSession(t *testing.T) {
	defaultProviderService = newProviderService()
	defaultProviderService.store.put(&providerSessionState{ID: "pending-session", Status: "pending", CreatedAt: defaultProviderService.now(), LastSeenAt: defaultProviderService.now()})
	t.Setenv("CSDN_PROVIDER_MODE", "real")
	server := newTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/articles?session=pending-session")
	if err != nil {
		t.Fatalf("articles request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 for pending session, got %d", resp.StatusCode)
	}

	defaultProviderService.store.put(&providerSessionState{ID: "authorized-session", Status: "authorized", CreatedAt: defaultProviderService.now(), LastSeenAt: defaultProviderService.now()})
	okResp, err := http.Get(server.URL + "/articles?session=authorized-session")
	if err != nil {
		t.Fatalf("authorized articles request failed: %v", err)
	}
	defer okResp.Body.Close()
	if okResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for authorized session, got %d", okResp.StatusCode)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	defaultProviderService = newProviderService()
	server := newTestServer()
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/login/start", nil)
	if err != nil {
		t.Fatalf("build request failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
