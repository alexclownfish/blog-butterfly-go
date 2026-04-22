package services

import (
	"strings"
	"testing"
)

func TestParseCSDNArticleExtractsStructuredContent(t *testing.T) {
	html := `<!doctype html><html><head>
<meta property="og:title" content="Go 并发实战" />
<meta name="description" content="讲清 goroutine 和 channel 的配合" />
<meta property="og:image" content="https://img.example.com/cover.png" />
<meta name="keywords" content="Go,goroutine,channel,CSDN,后端" />
</head><body>
<div class="article_content">
  <h1>Go 并发实战</h1>
  <p>第一段介绍。</p>
  <p>第二段包含 <strong>重点</strong> 和 <a href="https://go.dev">参考链接</a>。</p>
  <pre><code>fmt.Println("hi")
</code></pre>
  <img src="https://img.example.com/body.png" />
</div>
</body></html>`

	article, err := ParseCSDNArticle("https://blog.csdn.net/test/article/details/123", strings.NewReader(html))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if article.Title != "Go 并发实战" {
		t.Fatalf("expected title to be parsed, got %q", article.Title)
	}
	if article.Summary != "讲清 goroutine 和 channel 的配合" {
		t.Fatalf("expected summary from meta description, got %q", article.Summary)
	}
	if article.CoverImage != "https://img.example.com/cover.png" {
		t.Fatalf("expected cover image from og:image, got %q", article.CoverImage)
	}
	if article.Tags != "Go,goroutine,channel,后端" {
		t.Fatalf("expected cleaned tags, got %q", article.Tags)
	}
	if !strings.Contains(article.Content, "## Go 并发实战") {
		t.Fatalf("expected markdown heading in content, got %q", article.Content)
	}
	if !strings.Contains(article.Content, "**重点**") {
		t.Fatalf("expected bold markdown conversion, got %q", article.Content)
	}
	if !strings.Contains(article.Content, "[参考链接](https://go.dev)") {
		t.Fatalf("expected anchor markdown conversion, got %q", article.Content)
	}
	if !strings.Contains(article.Content, "```\nfmt.Println(\"hi\")\n```") {
		t.Fatalf("expected fenced code block, got %q", article.Content)
	}
	if !strings.Contains(article.Content, "![](https://img.example.com/body.png)") {
		t.Fatalf("expected image markdown conversion, got %q", article.Content)
	}
}

func TestValidateCSDNArticleURLRejectsNonCSDNHosts(t *testing.T) {
	if err := ValidateCSDNArticleURL("https://example.com/article/1"); err == nil {
		t.Fatal("expected non-CSDN host to be rejected")
	}
}
