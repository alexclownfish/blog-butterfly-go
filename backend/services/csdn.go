package services

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type CSDNArticle struct {
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Content        string `json:"content"`
	CoverImage     string `json:"cover_image"`
	Tags           string `json:"tags"`
	SourceURL      string `json:"source_url"`
	SourcePlatform string `json:"source_platform"`
}

var disallowedCSDNTags = map[string]struct{}{
	"csdn": {},
}

func ValidateCSDNArticleURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("解析链接失败: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("仅支持 http 或 https 链接")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" || !(host == "blog.csdn.net" || strings.HasSuffix(host, ".csdn.net")) {
		return errors.New("仅支持导入 CSDN 文章链接")
	}
	return nil
}

func FetchCSDNArticle(rawURL string) (*CSDNArticle, error) {
	if err := ValidateCSDNArticleURL(rawURL); err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; blog-butterfly-go/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("抓取 CSDN 文章失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("抓取 CSDN 文章失败，状态码: %d", resp.StatusCode)
	}

	return ParseCSDNArticle(rawURL, resp.Body)
}

func ParseCSDNArticle(rawURL string, r io.Reader) (*CSDNArticle, error) {
	if err := ValidateCSDNArticleURL(rawURL); err != nil {
		return nil, err
	}

	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	title := strings.TrimSpace(firstNonEmpty(
		findMetaContent(doc, "property", "og:title"),
		findMetaContent(doc, "name", "twitter:title"),
		findTitleText(doc),
	))
	if title == "" {
		return nil, errors.New("未能解析文章标题")
	}

	summary := strings.TrimSpace(firstNonEmpty(
		findMetaContent(doc, "name", "description"),
		findMetaContent(doc, "property", "og:description"),
	))
	cover := strings.TrimSpace(firstNonEmpty(
		findMetaContent(doc, "property", "og:image"),
		findMetaContent(doc, "name", "twitter:image"),
	))
	tags := normalizeTags(findMetaContent(doc, "name", "keywords"))

	contentNode := findArticleContentNode(doc)
	if contentNode == nil {
		return nil, errors.New("未找到文章正文")
	}
	markdown := strings.TrimSpace(renderMarkdown(contentNode))
	if markdown == "" {
		return nil, errors.New("文章正文为空")
	}
	if !strings.Contains(markdown, "## "+title) {
		markdown = "## " + title + "\n\n" + markdown
	}

	return &CSDNArticle{
		Title:          title,
		Summary:        summary,
		Content:        markdown,
		CoverImage:     cover,
		Tags:           tags,
		SourceURL:      strings.TrimSpace(rawURL),
		SourcePlatform: "csdn",
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func findMetaContent(root *html.Node, attrKey, attrValue string) string {
	var result string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil || result != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "meta" {
			if attrEquals(n, attrKey, attrValue) {
				result = attr(n, "content")
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return htmlEntityDecode(result)
}

func findTitleText(root *html.Node) string {
	var title string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil || title != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "title" {
			title = collectText(n)
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return strings.TrimSpace(strings.Split(title, "_CSDN")[0])
}

func findArticleContentNode(root *html.Node) *html.Node {
	selectors := []string{"article_content", "article-content", "blog-content-box", "articleContentId"}
	for _, selector := range selectors {
		if node := findNodeByClass(root, selector); node != nil {
			return node
		}
	}
	return nil
}

func findNodeByClass(root *html.Node, className string) *html.Node {
	var found *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil || found != nil {
			return
		}
		if n.Type == html.ElementNode {
			classAttr := attr(n, "class")
			for _, token := range strings.Fields(classAttr) {
				if token == className {
					found = n
					return
				}
			}
			if id := attr(n, "id"); id == className {
				found = n
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return found
}

func attr(node *html.Node, key string) string {
	for _, a := range node.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func attrEquals(node *html.Node, key, expected string) bool {
	return strings.EqualFold(strings.TrimSpace(attr(node, key)), expected)
}

func collectText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return htmlEntityDecode(node.Data)
	}
	var parts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text := strings.TrimSpace(collectText(child))
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func normalizeText(value string) string {
	value = htmlEntityDecode(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = regexp.MustCompile(`\s+`).ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func normalizeTags(raw string) string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '，', ';', '；', '|':
			return true
		default:
			return false
		}
	})
	seen := make(map[string]struct{})
	var cleaned []string
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, blocked := disallowedCSDNTags[strings.ToLower(tag)]; blocked {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		cleaned = append(cleaned, tag)
	}
	return strings.Join(cleaned, ",")
}

func renderMarkdown(node *html.Node) string {
	var b strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		renderNode(&b, child, 0)
	}
	return cleanupMarkdown(b.String())
}

func renderNode(b *strings.Builder, node *html.Node, listDepth int) {
	if node == nil {
		return
	}
	if node.Type == html.TextNode {
		text := normalizeText(node.Data)
		if text != "" {
			b.WriteString(text)
		}
		return
	}
	if node.Type != html.ElementNode {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			renderNode(b, child, listDepth)
		}
		return
	}

	switch node.Data {
	case "h1":
		writeBlock(b, "# "+inlineMarkdown(node))
	case "h2":
		writeBlock(b, "## "+inlineMarkdown(node))
	case "h3":
		writeBlock(b, "### "+inlineMarkdown(node))
	case "h4":
		writeBlock(b, "#### "+inlineMarkdown(node))
	case "h5":
		writeBlock(b, "##### "+inlineMarkdown(node))
	case "h6":
		writeBlock(b, "###### "+inlineMarkdown(node))
	case "p", "section", "blockquote", "div":
		content := inlineMarkdown(node)
		if content != "" {
			writeBlock(b, content)
		} else {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				renderNode(b, child, listDepth)
			}
		}
	case "pre":
		code := extractCodeText(node)
		if code != "" {
			writeBlock(b, "```\n"+code+"\n```")
		}
	case "ul":
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "li" {
				indent := strings.Repeat("  ", listDepth)
				item := inlineMarkdown(child)
				if item != "" {
					writeLine(b, indent+"- "+item)
				}
			}
		}
		b.WriteString("\n")
	case "ol":
		index := 1
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "li" {
				indent := strings.Repeat("  ", listDepth)
				item := inlineMarkdown(child)
				if item != "" {
					writeLine(b, fmt.Sprintf("%s%d. %s", indent, index, item))
					index++
				}
			}
		}
		b.WriteString("\n")
	case "img":
		src := strings.TrimSpace(attr(node, "src"))
		if src != "" {
			writeBlock(b, "![]("+src+")")
		}
	case "br":
		b.WriteString("\n")
	default:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			renderNode(b, child, listDepth)
		}
	}
}

func inlineMarkdown(node *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode {
			text := normalizeText(n.Data)
			if text != "" {
				if b.Len() > 0 {
					last := b.String()[b.Len()-1]
					if last != '\n' && last != ' ' && !strings.HasPrefix(text, ".") && !strings.HasPrefix(text, ",") && !strings.HasPrefix(text, "!") && !strings.HasPrefix(text, "?") && !strings.HasPrefix(text, ":") && !strings.HasPrefix(text, ";") {
						b.WriteByte(' ')
					}
				}
				b.WriteString(text)
			}
			return
		}
		if n.Type != html.ElementNode {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				walk(child)
			}
			return
		}

		switch n.Data {
		case "strong", "b":
			content := inlineMarkdownChildren(n)
			if content != "" {
				appendInline(&b, "**"+content+"**")
			}
		case "em", "i":
			content := inlineMarkdownChildren(n)
			if content != "" {
				appendInline(&b, "*"+content+"*")
			}
		case "code":
			content := normalizeText(collectText(n))
			if content != "" {
				appendInline(&b, "`"+content+"`")
			}
		case "a":
			text := inlineMarkdownChildren(n)
			href := strings.TrimSpace(attr(n, "href"))
			if text == "" {
				text = href
			}
			if href != "" {
				appendInline(&b, "["+text+"]("+href+")")
			} else if text != "" {
				appendInline(&b, text)
			}
		case "img":
			src := strings.TrimSpace(attr(n, "src"))
			if src != "" {
				appendInline(&b, "![]("+src+")")
			}
		case "br":
			appendInline(&b, "\n")
		default:
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				walk(child)
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child)
	}
	return strings.TrimSpace(b.String())
}

func inlineMarkdownChildren(node *html.Node) string {
	return strings.TrimSpace(inlineMarkdown(node))
}

func appendInline(b *strings.Builder, value string) {
	if value == "" {
		return
	}
	if b.Len() > 0 && !strings.HasPrefix(value, "\n") {
		last := b.String()[b.Len()-1]
		if last != ' ' && last != '\n' && !strings.HasPrefix(value, ".") && !strings.HasPrefix(value, ",") && !strings.HasPrefix(value, "!") && !strings.HasPrefix(value, "?") && !strings.HasPrefix(value, ":") && !strings.HasPrefix(value, ";") && !strings.HasPrefix(value, ")") {
			b.WriteByte(' ')
		}
	}
	b.WriteString(value)
}

func extractCodeText(node *html.Node) string {
	text := collectText(node)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.TrimSpace(text)
}

func writeBlock(b *strings.Builder, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	if b.Len() > 0 && !strings.HasSuffix(b.String(), "\n\n") {
		if strings.HasSuffix(b.String(), "\n") {
			b.WriteString("\n")
		} else {
			b.WriteString("\n\n")
		}
	}
	b.WriteString(content)
	b.WriteString("\n\n")
}

func writeLine(b *strings.Builder, content string) {
	if content == "" {
		return
	}
	b.WriteString(content)
	b.WriteString("\n")
}

func cleanupMarkdown(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	markdown = regexp.MustCompile(`\n{3,}`).ReplaceAllString(markdown, "\n\n")
	return strings.TrimSpace(markdown)
}

func htmlEntityDecode(value string) string {
	replacer := strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
	)
	return replacer.Replace(value)
}
