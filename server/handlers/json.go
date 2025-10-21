package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"pocketjson/storage"
	"pocketjson/utils"
)

func CreateJSON(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		ctx := r.Context()
		apiKey := r.Header.Get("X-API-Key")
		isAuth, _, err := store.ValidateApiKey(ctx, apiKey)
		if err != nil {
			log.Printf("api key validation error: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		cfg := store.Config()
		maxSize := cfg.DefaultMaxSize
		expiry := time.Now().Add(cfg.DefaultExpiry)
		creatorKey := "guest"
		var id string

		if isAuth {
			maxSize = cfg.AuthenticatedSize
			creatorKey = apiKey
			requestedID := chi.URLParam(r, "id")

			if requestedID != "" {
				if !utils.IsValidCustomID(requestedID) {
					http.Error(w, "Invalid ID format. Use only alphanumeric characters, hyphens, and underscores (max 64 chars)", http.StatusBadRequest)
					return
				}
				clientPrefix := utils.GetClientPrefix(apiKey)
				id = fmt.Sprintf("%s_%s", clientPrefix, requestedID)
			} else {
				var err error
				id, err = utils.GenerateRandomKey()
				if err != nil {
					log.Printf("failed to generate random key: %v", err)
					http.Error(w, "Server error", http.StatusInternalServerError)
					return
				}
			}

			if exp := r.URL.Query().Get("expiry"); exp != "" {
				if exp == "never" {
					expiry = time.Now().AddDate(100, 0, 0)
				} else if hours, err := strconv.Atoi(exp); err == nil {
					expiry = time.Now().Add(time.Duration(hours) * time.Hour)
				}
			}
		} else {
			var err error
			id, err = utils.GenerateRandomKey()
			if err != nil {
				log.Printf("failed to generate random key: %v", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
		}

		if len(jsonBytes) > maxSize {
			http.Error(w, "JSON too large", http.StatusBadRequest)
			return
		}

		if err := store.DB().CreateJSON(ctx, id, string(jsonBytes), expiry, creatorKey); err != nil {
			log.Printf("failed to store JSON: %v", err)
			http.Error(w, "Failed to store JSON", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         id,
			"expires_at": expiry.Format(time.RFC3339),
		})
	}
}

func GetJSON(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		ctx := r.Context()

		data, err := store.DB().GetJSON(ctx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "JSON not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to retrieve JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(data))
	}
}
