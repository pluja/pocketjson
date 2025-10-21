package storage

import (
	"context"
	"crypto/subtle"
	"log"
	"sync"
	"time"

	"pocketjson/config"
)

type apiKeyCacheEntry struct {
	isValid bool
	isAdmin bool
	expires time.Time
}

type Store struct {
	db           *DB
	config       *config.Config
	cleanup      sync.WaitGroup
	ctx          context.Context
	cancelCtx    context.CancelFunc
	apiKeyCache  map[string]apiKeyCacheEntry
	cacheMutex   sync.RWMutex
	cacheTTL     time.Duration
}

func New(db *DB, cfg *config.Config) *Store {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Store{
		db:          db,
		config:      cfg,
		ctx:         ctx,
		cancelCtx:   cancel,
		apiKeyCache: make(map[string]apiKeyCacheEntry),
		cacheTTL:    5 * time.Minute,
	}
	s.startCleanupRoutine()
	s.startCacheCleanupRoutine()
	return s
}

func (s *Store) startCleanupRoutine() {
	s.cleanup.Add(1)
	go func() {
		defer s.cleanup.Done()
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
				deleted, err := s.db.DeleteExpiredJSON(ctx)
				cancel()
				if err != nil {
					log.Printf("cleanup error: %v", err)
				} else if deleted > 0 {
					log.Printf("cleanup: deleted %d expired entries", deleted)
				}
			}
		}
	}()
}

func (s *Store) startCacheCleanupRoutine() {
	s.cleanup.Add(1)
	go func() {
		defer s.cleanup.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.cleanExpiredCacheEntries()
			}
		}
	}()
}

func (s *Store) cleanExpiredCacheEntries() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	now := time.Now()
	for key, entry := range s.apiKeyCache {
		if now.After(entry.expires) {
			delete(s.apiKeyCache, key)
		}
	}
}

// ValidateApiKey validates an API key and returns (isValid, isAdmin, error).
// Uses constant-time comparison for master key and caches results for performance.
// Returns an error only if database operations fail; authentication failures return (false, false, nil).
func (s *Store) ValidateApiKey(ctx context.Context, key string) (bool, bool, error) {
	if len(key) == len(s.config.MasterAPIKey) &&
		subtle.ConstantTimeCompare([]byte(key), []byte(s.config.MasterAPIKey)) == 1 {
		return true, true, nil
	}

	if key == "" {
		return false, false, nil
	}

	s.cacheMutex.RLock()
	if cached, found := s.apiKeyCache[key]; found && time.Now().Before(cached.expires) {
		s.cacheMutex.RUnlock()
		return cached.isValid, cached.isAdmin, nil
	}
	s.cacheMutex.RUnlock()

	isAdmin, _, err := s.db.GetApiKey(ctx, key)
	if err != nil {
		if err.Error() == "api key not found" {
			s.cacheApiKey(key, false, false, 30*time.Second)
			return false, false, nil
		}
		log.Printf("api key validation error: %v", err)
		return false, false, err
	}

	s.cacheApiKey(key, true, isAdmin, s.cacheTTL)
	return true, isAdmin, nil
}

func (s *Store) cacheApiKey(key string, isValid, isAdmin bool, ttl time.Duration) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.apiKeyCache[key] = apiKeyCacheEntry{
		isValid: isValid,
		isAdmin: isAdmin,
		expires: time.Now().Add(ttl),
	}
}

func (s *Store) InvalidateApiKeyCache(key string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	delete(s.apiKeyCache, key)
}

func (s *Store) DB() *DB {
	return s.db
}

func (s *Store) Config() *config.Config {
	return s.config
}

func (s *Store) Shutdown() {
	s.cancelCtx()
	s.cleanup.Wait()
}
