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
- ✅ Rate-limiting and middleware support
- ✅ Fully tested with 90%+ coverage

---

## 🏗️ Project Structure

.
├── internal/analyzer/   # Core analysis logic
├── cmd/server/          # HTTP server entry point
├── config/              # Config for tag extraction
├── templates/           # HTML templates for UI
├── static/              # Static assets (JS, CSS, etc.)
├── Dockerfile           # Docker container instructions
├── Makefile             # Dev and CI tasks
└── go.mod / go.sum      # Module dependencies

---

## 🧪 Running Tests

Run tests and generate coverage:

```bash
 make test
 make cover
```
⸻

🔧 Running Locally

Start the application:

```bash
go run ./cmd/server
```

Visit: http://localhost:8080

⸻

🐳 Docker Usage

Build and run with Docker:

docker build -t webanalyzer .
docker run -p 8080:8080 webanalyzer

Or use Make:

make docker
make run-docker


⸻

🔌 API Usage

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

Customize heading tags in config/config.json:

```bash
{
  "headings": ["article", "section", "summary"]
}
```

⸻

🧰 Developer Tools
	•	Go 1.22+
	•	Bootstrap 5
	•	golang.org/x/net/html
	•	Custom rate-limiting middleware
	•	Test coverage: 90%+

⸻

📜 License

MIT License – feel free to use and modify.

