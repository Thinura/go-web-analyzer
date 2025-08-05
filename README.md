# 🌐 Web Page Analyzer

A Go-powered web application that analyzes any public webpage and gives you insights like HTML version, page title, heading structure, internal/external links, accessibility of links, and more.

Designed for developers, testers, SEO professionals, and curious minds.

---

## 🚀 Features

- ✅ Analyze HTML version (HTML5, XHTML, etc.)
- ✅ Extract and count headings (h1–h6 and custom tags)
- ✅ Detect login forms
- ✅ Identify internal and external links
- ✅ Check link accessibility using concurrent HTTP requests
- ✅ Categorize links as accessible/inaccessible
- ✅ Measure analysis time
- ✅ JSON API endpoint for integration
- ✅ Beautiful Bootstrap UI dashboard
- ✅ Render JS-heavy pages using Puppeteer
- ✅ Rate-limiting and middleware
- ✅ 90%+ test coverage

---

## 🏗️ Project Structure

```text
.
web-analyzer/
├── cmd/
│   └── webanalyzer/            # Entry point (main.go)
├── internal/
│   ├── analyzer/               # Core logic (analysis, config, fetchers)
│   ├── constants/              # Constants shared within internal
│   ├── helpers/                # Utility fetchers (TryStandard, etc.)
│   └── server/                 # Handlers and middleware
├── pkg/
│   ├── configloader/           # External config reading logic
│   ├── domrenderer/            # Puppeteer integration
│   ├── embed/                  # go:embed usage
│   │   ├── config/
│   │   │   └── config.json     # JSON config for custom tags
│   │   └── templates/          # HTML templates
│   └── errors/                 # Custom HTTP error types
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod / go.sum
```

---

## 🔧 Running Locally

```bash
 cd web-analyzer
 go run ./cmd/webanalyzer 
```

Visit: http://localhost:8080

---

## 🧪 Running Tests

Run tests and generate coverage:

```bash
 cd web-analyzer
 go test -coverprofile=coverage.out -count=1 ./...
```

---

## 🐳 Docker Usage

Build and run with Docker:

```bash
docker-compose up --build
```

---

## 🔌 API Usage

Send a POST request to analyze a page:

POST /api/analyze
Content-Type: application/x-www-form-urlencoded

Request Body:

url=https://example.com

Response:
```bash
{
  "PageURL": "https://example.com",
  "HTMLVersion": "HTML5",
  "Title": "Example Domain",
  "Headings": [
    {"Tag": "h1", "Title": "Example Heading"}
  ],
  "InternalLinks": [...],
  "ExternalLinks": [...],
  "AccessibleLinks": [...],
  "InaccessibleLinks": [...],
  "HasLoginForm": false,
  "AnalysisDuration": "1.23s"
}
```

⸻

⚙️ Configuration

Customize heading tags in pkg/embed/config/config.json:

```bash
{
  "headings": ["h1", "h2", "h3"]
}
```

⸻

🧰 Developer Tools
- Go 1.22+
- Bootstrap 5
- Puppeteer / Playwright
- golang.org/x/net/html
- Docker + Compose
- Custom middleware & rate limiter
- go:embed for HTML/config embedding
- Test Coverage: 75%+ 

⸻

📜 License

MIT License – feel free to use and modify.

