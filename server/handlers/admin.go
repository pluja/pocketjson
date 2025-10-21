package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"pocketjson/storage"
	"pocketjson/utils"
)

func AdminOnly(store *storage.Store) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			isAuth, isAdmin, err := store.ValidateApiKey(r.Context(), apiKey)

			if err != nil {
				log.Printf("admin auth error: %v", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}

			if !isAuth || !isAdmin {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
	}
}

func CreateApiKey(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Description string `json:"description"`
			IsAdmin     bool   `json:"is_admin"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		key, err := utils.GenerateRandomKey()
		if err != nil {
			log.Printf("failed to generate API key: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		if err := store.DB().CreateApiKey(r.Context(), key, request.Description, request.IsAdmin); err != nil {
			log.Printf("failed to create API key: %v", err)
			http.Error(w, "Failed to create API key", http.StatusInternalServerError)
			return
		}

		clientId := utils.GetClientPrefix(key)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"key":         key,
			"client_id":   clientId,
			"description": request.Description,
			"is_admin":    request.IsAdmin,
			"created_at":  time.Now().Format(time.RFC3339),
		})
	}
}

func DeleteApiKey(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyToDelete := chi.URLParam(r, "key")

		if err := store.DB().DeleteApiKey(r.Context(), keyToDelete); err != nil {
			log.Printf("failed to delete API key: %v", err)
			http.Error(w, "Failed to delete API key", http.StatusInternalServerError)
			return
		}

		store.InvalidateApiKeyCache(keyToDelete)

		w.WriteHeader(http.StatusOK)
	}
}
