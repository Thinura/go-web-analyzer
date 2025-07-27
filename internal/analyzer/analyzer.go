package analyzer

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type NamedLink struct {
	URL   string
	Label string
}

type Heading struct {
	Tag   string
	Title string
}

type Result struct {
	PageURL           string
	HTMLVersion       string
	Title             string
	Headings          []Heading
	HeadingsCount     map[string]int
	InternalLinks     []NamedLink
	ExternalLinks     []NamedLink
	AccessibleLinks   []NamedLink
	InaccessibleLinks []NamedLink
	HasLoginForm      bool
	AnalysisDuration  time.Duration
}

type LinkCheckerConfig struct {
	MaxConcurrency int
	Timeout        time.Duration
	Logger         func(format string, args ...interface{}) // nil = silent
}

func stripPort(hostport string) string {
	host := hostport
	if colon := strings.Index(hostport, ":"); colon != -1 {
		host = hostport[:colon]
	}
	return host
}

func deduplicateLinks(links []string) []NamedLink {
	seen := make(map[string]int)
	var named []NamedLink
	for _, link := range links {
		seen[link]++
		label := link
		if seen[link] > 1 {
			label = fmt.Sprintf("%s (duplicate %d)", link, seen[link])
		}
		named = append(named, NamedLink{URL: link, Label: label})
	}
	return named
}

// AnalyzePage analyzes the given URL.
func AnalyzePage(pageURL string) (*Result, error) {
	start := time.Now()
	parsedURL, err := url.ParseRequestURI(pageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	htmlVersion := detectHTMLVersion(data)

	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	result := &Result{
		PageURL:       pageURL,
		HeadingsCount: make(map[string]int),
		HTMLVersion:   htmlVersion,
	}

	extractInfo(doc, parsedURL, result)

	result.AnalysisDuration = time.Since(start)
	return result, nil
}

func detectHTMLVersion(data []byte) string {
	// Check only the first 256 bytes
	snippet := strings.ToLower(string(bytes.TrimSpace(data)))
	if len(snippet) > 256 {
		snippet = snippet[:256]
	}

	// Known version checks
	switch {
	case strings.Contains(snippet, "<!doctype html>"):
		return "HTML5"
	case strings.Contains(snippet, "html 4.01 transitional"):
		return "HTML 4.01 Transitional"
	case strings.Contains(snippet, "html 4.01//en"):
		return "HTML 4.01 Strict"
	case strings.Contains(snippet, "xhtml 1.0 strict"):
		return "XHTML 1.0 Strict"
	case strings.Contains(snippet, "xhtml 1.0 transitional"):
		return "XHTML 1.0 Transitional"
	}

	re := regexp.MustCompile(`(?i)<!doctype\s+([^>]+)>`)
	if matches := re.FindStringSubmatch(snippet); len(matches) > 1 {
		return fmt.Sprintf("Unknown DOCTYPE: %s", matches[1])
	}

	return "Unknown or Custom DOCTYPE"
}

func contains(list []string, key string) bool {
	for _, item := range list {
		if item == key {
			return true
		}
	}
	return false
}

func getTextContent(n *html.Node) string {
	var builder strings.Builder
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			builder.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(n)
	return builder.String()
}

// Walk the DOM and extract info.
func extractInfo(n *html.Node, baseURL *url.URL, result *Result) {
	var rawInternal []string
	var rawExternal []string
	var allLinks []string

	cfg, err := LoadTagConfig("config/config.json")
	if err != nil {
		log.Println("Failed to load config:", err)
	} else {
		log.Println("Loaded headings config:", cfg.Headings)
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil {
					result.Title = n.FirstChild.Data
				}
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						linkURL, err := url.Parse(attr.Val)
						if err != nil || attr.Val == "" {
							continue
						}
						resolved := baseURL.ResolveReference(linkURL)
						full := resolved.String()
						allLinks = append(allLinks, full)

						baseHost := strings.ToLower(stripPort(baseURL.Host))
						linkHost := strings.ToLower(stripPort(resolved.Host))
						if baseHost == linkHost {
							rawInternal = append(rawInternal, full)
						} else {
							rawExternal = append(rawExternal, full)
						}
					}
				}
			case "input":
				for _, attr := range n.Attr {
					if attr.Key == "type" && strings.ToLower(attr.Val) == "password" {
						result.HasLoginForm = true
					}
				}
			default:
				if contains(cfg.Headings, n.Data) {
					result.HeadingsCount[n.Data]++
					result.Headings = append(result.Headings, Heading{
						Tag:   n.Data,
						Title: strings.TrimSpace(getTextContent(n)),
					})
				}

			}

		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	result.InternalLinks = deduplicateLinks(rawInternal)
	result.ExternalLinks = deduplicateLinks(rawExternal)
}

func checkLinksConcurrently(links []string, config LinkCheckerConfig) int {
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.MaxConcurrency)
	bad := make(chan bool, len(links))

	for _, link := range links {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if !isLinkAccessible(url, config.Timeout, config.Logger) {
				bad <- true
			}
		}(link)
	}

	wg.Wait()
	close(bad)

	count := 0
	for b := range bad {
		if b {
			count++
		}
	}
	return count
}

func isLinkAccessible(link string, timeout time.Duration, logger func(string, ...interface{})) bool {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		if logger != nil {
			logger("HEAD request creation failed for %s: %v", link, err)
		}
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		if logger != nil {
			logger("HEAD request failed for %s: %v", link, err)
		}
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func ClassifyLinksConcurrently(links []NamedLink, config LinkCheckerConfig) (accessible, inaccessible []NamedLink) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.MaxConcurrency)
	mu := sync.Mutex{}

	// Track duplicates
	nameCount := make(map[string]int)
	uniqueLinks := make([]NamedLink, len(links))
	for i, l := range links {
		nameCount[l.Label]++
		count := nameCount[l.Label]
		if count > 1 {
			l.Label = fmt.Sprintf("%s (duplicate %d)", l.Label, count)
		}
		uniqueLinks[i] = l
	}

	for _, link := range uniqueLinks {
		wg.Add(1)
		go func(link NamedLink) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ok := isLinkAccessible(link.URL, config.Timeout, config.Logger)

			mu.Lock()
			if ok {
				accessible = append(accessible, link)
			} else {
				inaccessible = append(inaccessible, link)
			}
			mu.Unlock()
		}(link)
	}

	wg.Wait()
	return
}
