package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"pocketjson/storage"
)

var homeTemplate *template.Template

func init() {
	possiblePaths := []string{
		"templates/home.html",
		"../../templates/home.html",
		filepath.Join(".", "templates", "home.html"),
	}

	var err error
	for _, path := range possiblePaths {
		if _, statErr := os.Stat(path); statErr == nil {
			homeTemplate, err = template.ParseFiles(path)
			if err == nil {
				return
			}
		}
	}

	homeTemplate = template.Must(template.New("home").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>PocketJSON</title>
</head>
<body>
    <h1>PocketJSON Storage Service</h1>
    <p>Welcome to PocketJSON, a lightweight, single-binary JSON storage service with built-in expiry and multi-tenant support. Perfect for developers who need a quick, reliable way to store and retrieve JSON data without the overhead of a full database setup.</p>

    {{if .InstanceInfo}}
        {{.InstanceInfo}}
    {{else}}
    <ul>
        <li>Read the <a href="https://github.com/pluja/pocketjson?tab=readme-ov-file#api-reference-">API Docs</a></li>
        <li>No backups. If your data is lost due to some technical issues, its lost forever.</li>
        <li>Maximum allowed payload size cannot be more than {{.MaxSizeKB}} Kb per request for guest users.</li>
        <li>Guest users expiration time is {{.ExpiryHours}} hours</li>
        <li>Guest rate limit of {{.RateLimit}} req/min</li>
        <li>This is meant for small projects and that's why it is offered FREE of cost.</li>
    </ul>
    {{end}}

    <p><a href="https://github.com/pluja/pocketjson#readme">Source Code</a></p>
</body>
</html>`))
}

func ServeHomePage(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := store.Config()

		data := struct {
			InstanceInfo string
			MaxSizeKB    int
			ExpiryHours  int
			RateLimit    int
		}{
			InstanceInfo: os.Getenv("INSTANCE_INFO"),
			MaxSizeKB:    cfg.DefaultMaxSize / 1024,
			ExpiryHours:  int(cfg.DefaultExpiry.Hours()),
			RateLimit:    cfg.RequestLimit,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := homeTemplate.Execute(w, data); err != nil {
			log.Printf("failed to execute template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
