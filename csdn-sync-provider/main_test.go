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

func TestMethodNotAllowed(t *testing.T) {
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
