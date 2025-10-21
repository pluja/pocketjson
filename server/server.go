package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"pocketjson/server/handlers"
	custommw "pocketjson/server/middleware"
	"pocketjson/storage"
)

type Server struct {
	store  *storage.Store
	router *chi.Mux
	server *http.Server
}

func New(store *storage.Store) *Server {
	s := &Server{
		store:  store,
		router: chi.NewRouter(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	cfg := store.Config()
	s.server = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: s.router,
	}

	return s
}

func (s *Server) setupMiddleware() {
	cfg := s.store.Config()

	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	s.router.Use(custommw.RateLimit(s.store))
}

func (s *Server) setupRoutes() {
	adminOnly := handlers.AdminOnly(s.store)

	s.router.Get("/health", handlers.HealthCheck)
	s.router.Get("/", handlers.ServeHomePage(s.store))

	s.router.Post("/", handlers.CreateJSON(s.store))
	s.router.Post("/{id}", handlers.CreateJSON(s.store))
	s.router.Get("/{id}", handlers.GetJSON(s.store))

	s.router.Post("/admin/keys", adminOnly(handlers.CreateApiKey(s.store)))
	s.router.Delete("/admin/keys/{key}", adminOnly(handlers.DeleteApiKey(s.store)))
}

func (s *Server) Start() error {
	cfg := s.store.Config()
	log.Printf("Server starting on :%s", cfg.Port)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.store.Shutdown()

	log.Println("Server stopped gracefully")
	return nil
}
