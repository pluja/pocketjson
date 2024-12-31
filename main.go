package main

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	_ "github.com/mattn/go-sqlite3"

	"pocketjson/ent"
	"pocketjson/ent/apikey"
	"pocketjson/ent/jsonstorage"
)

var (
	masterApiKey      = ""
	defaultMaxSize    = getEnvInt("DEFAULT_MAX_SIZE", 100*1024)        // 100KB default
	authenticatedSize = getEnvInt("AUTHENTICATED_MAX_SIZE", 1024*1024) // 1MB default
	defaultExpiry     = time.Duration(getEnvInt("DEFAULT_EXPIRY_HOURS", 48)) * time.Hour
	requestLimit      = getEnvInt("REQUEST_LIMIT", 15)
)

type JsonStore struct {
	client  *ent.Client
	cleanup sync.WaitGroup
}

func getEnvStr(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func generateRandomKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func NewJsonStore(client *ent.Client) *JsonStore {
	js := &JsonStore{client: client}
	js.startCleanupRoutine()
	return js
}

func (js *JsonStore) startCleanupRoutine() {
	js.cleanup.Add(1)
	go func() {
		defer js.cleanup.Done()
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			_, err := js.client.JsonStorage.Delete().
				Where(jsonstorage.ExpiresAtLT(time.Now())).
				Exec(ctx)
			if err != nil {
				log.Printf("cleanup error: %v", err)
			}
		}
	}()
}

func (js *JsonStore) validateApiKey(ctx context.Context, key string) (bool, bool, error) {
	if key == masterApiKey {
		return true, true, nil
	}

	if key == "" {
		return false, false, nil
	}

	apiKey, err := js.client.ApiKey.Query().
		Where(apikey.Key(key)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return false, false, nil
		}
		return false, false, err
	}

	return true, apiKey.IsAdmin, nil
}

func getClientPrefix(apiKey string) string {
	hash := md5.Sum([]byte(apiKey))
	return fmt.Sprintf("%x", hash)[:5]
}

func (js *JsonStore) CreateJSON(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to process JSON", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	apiKey := r.Header.Get("X-API-Key")
	isAuth, _, err := js.validateApiKey(ctx, apiKey)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	maxSize := defaultMaxSize
	expiry := time.Now().Add(defaultExpiry)
	creatorKey := "guest"
	var id string

	if isAuth {
		maxSize = authenticatedSize
		creatorKey = apiKey
		requestedID := chi.URLParam(r, "id")

		if requestedID != "" {
			clientPrefix := getClientPrefix(apiKey)
			id = fmt.Sprintf("%s_%s", clientPrefix, requestedID)
		} else {
			id = generateRandomKey()
		}

		if exp := r.URL.Query().Get("expiry"); exp != "" {
			if exp == "never" {
				expiry = time.Now().AddDate(100, 0, 0)
			} else if hours, err := strconv.Atoi(exp); err == nil {
				expiry = time.Now().Add(time.Duration(hours) * time.Hour)
			}
		}
	} else {
		id = generateRandomKey()
	}

	if len(jsonBytes) > maxSize {
		http.Error(w, "JSON too large", http.StatusBadRequest)
		return
	}

	storage, err := js.client.JsonStorage.Create().
		SetID(id).
		SetData(string(jsonBytes)).
		SetExpiresAt(expiry).
		SetCreatorKey(creatorKey).
		Save(ctx)

	if err != nil {
		http.Error(w, "Failed to store JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         storage.ID,
		"expires_at": storage.ExpiresAt.Format(time.RFC3339),
	})
}

func (js *JsonStore) GetJSON(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := context.Background()

	storage, err := js.client.JsonStorage.Query().
		Where(jsonstorage.ID(id)).
		Where(jsonstorage.ExpiresAtGT(time.Now())).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, "JSON not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(storage.Data))
}

func (js *JsonStore) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		apiKey := r.Header.Get("X-API-Key")
		isAuth, isAdmin, err := js.validateApiKey(ctx, apiKey)

		if err != nil || !isAuth || !isAdmin {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func (js *JsonStore) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Description string `json:"description"`
		IsAdmin     bool   `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	key := generateRandomKey()
	ctx := context.Background()

	apiKey, err := js.client.ApiKey.Create().
		SetKey(key).
		SetDescription(request.Description).
		SetIsAdmin(request.IsAdmin).
		Save(ctx)

	if err != nil {
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	clientId := getClientPrefix(apiKey.Key)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":         apiKey.Key,
		"client_id":   clientId,
		"description": apiKey.Description,
		"is_admin":    apiKey.IsAdmin,
		"created_at":  apiKey.CreatedAt,
	})
}

func (js *JsonStore) DeleteApiKey(w http.ResponseWriter, r *http.Request) {
	keyToDelete := chi.URLParam(r, "key")
	ctx := context.Background()

	_, err := js.client.ApiKey.Delete().
		Where(apikey.Key(keyToDelete)).
		Exec(ctx)

	if err != nil {
		http.Error(w, "Failed to delete API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func serveHomePage(w http.ResponseWriter, r *http.Request) {
	instanceInfo := os.Getenv("INSTANCE_INFO")
	if instanceInfo == "" {
		instanceInfo = fmt.Sprintf(`
    <ul>
		<li>Read the <a href="https://github.com/pluja/pocketjson?tab=readme-ov-file#api-reference-">API Docs</a></li>
		<li>No backups. If your data is lost due to some technical issues, its lost forever.</li>
		<li>Maximum allowed payload size cannot be more than %d Kb per request for guest users.</li>
		<li>Guest users expiration time is %d hours</li>
		<li>Guest rate limit of %d req/min</li>
		<li>This is meant for small projects and that's why it is offered FREE of cost.</li>
	</ul>`, defaultMaxSize/1024, int(defaultExpiry.Hours()), requestLimit)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>PocketJSON</title>
</head>
<body>
    <h1>PocketJSON Storage Service</h1>
    <p>Welcome to PocketJSON, a lightweight, single-binary JSON storage service with built-in expiry and multi-tenant support. Perfect for developers who need a quick, reliable way to store and retrieve JSON data without the overhead of a full database setup.</p>
    %s
    <p><a href="https://github.com/pluja/pocketjson#readme">Source Code</a></p>
</body>
</html>
    `, instanceInfo)))
}

func main() {
	dataDir := "data"
	os.MkdirAll(dataDir, 0755)

	client, err := ent.Open("sqlite3", filepath.Join(dataDir, "jsonstore.db")+"?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	js := NewJsonStore(client)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{getEnvStr("CORS_ALLOWED_ORIGINS", "*")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(func(next http.Handler) http.Handler {
		limiter := httprate.Limit(
			requestLimit,
			1*time.Minute,
			httprate.WithKeyFuncs(httprate.KeyByIP),
		)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			isAuth, _, err := js.validateApiKey(r.Context(), apiKey)
			if err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}

			if !isAuth {
				limiter(next).ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	r.Get("/health", healthCheck)
	r.Get("/", serveHomePage)
	r.Post("/", js.CreateJSON)
	r.Post("/{id}", js.CreateJSON)
	r.Get("/{id}", js.GetJSON)
	r.Post("/admin/keys", js.adminOnly(js.CreateApiKey))
	r.Delete("/admin/keys/{key}", js.adminOnly(js.DeleteApiKey))

	// Generate a random key if no master key is provided
	masterApiKey = getEnvStr("MASTER_API_KEY", "")
	if masterApiKey == "" {
		masterApiKey = generateRandomKey()
		log.Printf("WARNING: No master API key provided. Generated random key: %s", masterApiKey)
		log.Println("Please save this key and set it as MASTER_API_KEY environment variable for subsequent runs")
	}

	server := &http.Server{
		Addr:    ":9819",
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		js.cleanup.Wait()
		os.Exit(0)
	}()

	log.Println("Server starting on :9819")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
