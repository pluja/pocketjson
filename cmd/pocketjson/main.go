package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"pocketjson/config"
	"pocketjson/server"
	"pocketjson/storage"
	"pocketjson/utils"
)

func main() {
	cfg := config.Load()

	if cfg.MasterAPIKey == "" {
		key, err := utils.GenerateRandomKey()
		if err != nil {
			log.Fatalf("failed to generate master API key: %v", err)
		}
		cfg.MasterAPIKey = key
		log.Printf("WARNING: No master API key provided. Generated random key: %s", key)
		log.Println("Please save this key and set it as MASTER_API_KEY environment variable for subsequent runs")
	}

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// SQLite optimizations: WAL mode for concurrency, 32MB cache, 5s busy timeout
	dbPath := filepath.Join(cfg.DataDir, "jsonstore.db") + "?_fk=1&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&cache_size=-32000"
	db, err := storage.NewDB(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	store := storage.New(db, cfg)
	srv := server.New(store)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
		os.Exit(1)
	}
}
