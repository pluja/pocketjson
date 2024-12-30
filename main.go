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
	"github.com/go-chi/httprate"
	_ "github.com/mattn/go-sqlite3"

	"pocketjson/ent"
	"pocketjson/ent/apikey"
	"pocketjson/ent/jsonstorage"
)

const (
	defaultMaxSize    = 100 * 1024  // 100KB
	authenticatedSize = 1024 * 1024 // 1MB
	defaultExpiry     = 24 * time.Hour
)

var (
	masterApiKey = "master-key-replace-in-production" // In production, use environment variables
)

type JsonStore struct {
	client  *ent.Client
	cleanup sync.WaitGroup
}

func generateRandomKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func NewJsonStore(client *ent.Client) *JsonStore {
	js := &JsonStore{client: client}
	// Start the cleanup routine
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
			// Delete expired entries
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

// Add this helper function to get client prefix from API key
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
		clientPrefix := getClientPrefix(apiKey)
		id = fmt.Sprintf("%s_%s", clientPrefix, requestedID)

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

	// Generate the client ID from the API key
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

func main() {
	// Ensure data directory exists
	dataDir := "data"
	os.MkdirAll(dataDir, 0755)

	// Modified connection string to enable foreign keys
	client, err := ent.Open("sqlite3", filepath.Join(dataDir, "jsonstore.db")+"?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	js := NewJsonStore(client)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Apply rate limiting only to non-authenticated requests
	r.Use(func(next http.Handler) http.Handler {
		limiter := httprate.Limit(
			15,                                      // requests
			1*time.Minute,                           // per minute
			httprate.WithKeyFuncs(httprate.KeyByIP), // rate limit by IP
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

	// JSON storage endpoints
	r.Post("/{id}", js.CreateJSON)
	r.Get("/{id}", js.GetJSON)

	// API key management endpoints (admin only)
	r.Post("/admin/keys", js.adminOnly(js.CreateApiKey))
	r.Delete("/admin/keys/{key}", js.adminOnly(js.DeleteApiKey))

	masterApiKey = os.Getenv("MASTER_API_KEY")
	if masterApiKey == "" {
		masterApiKey = "master-key-replace-in-production"
		log.Println("Warning: Using default master API key. Consider setting MASTER_API_KEY environment variable.")
	}

	server := &http.Server{
		Addr:    ":9819",
		Handler: r,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown the server
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		// Wait for cleanup routine to finish
		js.cleanup.Wait()
		os.Exit(0)
	}()

	log.Println("Server starting on :9819")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
