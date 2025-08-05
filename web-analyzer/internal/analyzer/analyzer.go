package analyzer

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"web-analyzer/internal/helpers"
	"web-analyzer/pkg/errors"

	"golang.org/x/net/html"
)

type NamedLink struct {
	URL        string
	Label      string
	Occurrence int
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

// AnalyzePage analyzes the given URL.
// func AnalyzePage(pageURL string) (*Result, error) {
// 	start := time.Now()
// 	parsedURL, err := url.ParseRequestURI(pageURL)
// 	if err != nil {
// 		return nil, &errors.HTTPError{StatusCode: http.StatusBadRequest, Message: fmt.Sprintf("invalid URL: %v", err)}
// 	}

// 	client := &http.Client{Timeout: 10 * time.Second}
// 	resp, err := client.Get(pageURL)
// 	if err != nil {
// 		return nil, &errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("failed to fetch: %v", err)}
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, &errors.HTTPError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status)}
// 	}

// 	data, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, &errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("failed to read response body: %v", err)}
// 	}

// 	htmlVersion := detectHTMLVersion(data)

// 	doc, err := html.Parse(strings.NewReader(string(data)))
// 	if err != nil {
// 		return nil, &errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("failed to parse HTML: %v", err)}
// 	}

// 	result := &Result{
// 		PageURL:     pageURL,
// 		HTMLVersion: htmlVersion,
// 	}

// 	extractInfo(doc, parsedURL, result)

// 	result.AnalysisDuration = time.Since(start)
// 	return result, nil
// }

func AnalyzePage(pageURL string) (*Result, error) {
	start := time.Now()
	parsedURL, err := url.ParseRequestURI(pageURL)
	if err != nil {
		return nil, &errors.HTTPError{StatusCode: http.StatusBadRequest, Message: fmt.Sprintf("invalid URL: %v", err)}
	}

	data, isBotBlocked, err := helpers.TryStandardFetch(pageURL)
	if err != nil {
		return nil, err
	}

	// Retry with Puppeteer render if bot-block detected
	if isBotBlocked {
		rendered, err := helpers.FetchRenderedDOM(pageURL)
		if err != nil {
			return nil, &errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("puppeteer render failed: %v", err)}
		}
		data = rendered
	}

	htmlVersion := detectHTMLVersion(data)
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, &errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("failed to parse HTML: %v", err)}
	}

	result := &Result{
		PageURL:     pageURL,
		HTMLVersion: htmlVersion,
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

	cfg, err := LoadTagConfig()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		panic(&errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("Failed to load config: %v", err)})
	}
	log.Println("Loaded headings config:", cfg.Headings)

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

	result.InternalLinks = ToNamedLinks(rawInternal)
	result.ExternalLinks = ToNamedLinks(rawExternal)
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

	for _, link := range links {
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

func ToNamedLinks(links []string) []NamedLink {
	countMap := make(map[string]int)

	for _, link := range links {
		countMap[link]++
	}

	named := make([]NamedLink, 0, len(countMap))
	for url, count := range countMap {
		named = append(named, NamedLink{
			URL:        url,
			Label:      url,
			Occurrence: count,
		})
	}

	return named
}

func RelabelDuplicates(links []NamedLink) []NamedLink {
	seen := make(map[string]int)

	// Count occurrences
	for _, l := range links {
		seen[l.URL]++
	}

	// Create final labeled slice
	var labeled []NamedLink
	for url, count := range seen {
		label := url
		if count > 1 {
			label = fmt.Sprintf("%s (%d)", url, count)
		}
		labeled = append(labeled, NamedLink{URL: url, Label: label})
	}

	return labeled
}
