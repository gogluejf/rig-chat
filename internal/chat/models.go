package chat

import (
	"context"
	"sync"
	"time"

	"rig-chat/internal/config"
)

// ModelEntry represents a discovered model with its provider
type ModelEntry struct {
	ID       string
	Provider string
}

// ModelCache caches discovered models with a TTL
type ModelCache struct {
	mu       sync.RWMutex
	models   []ModelEntry
	cachedAt time.Time
	ttl      time.Duration
}

func NewModelCache(ttl time.Duration) *ModelCache {
	return &ModelCache{ttl: ttl}
}

// Get returns cached models if still valid
func (mc *ModelCache) Get() ([]ModelEntry, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if mc.models == nil || time.Since(mc.cachedAt) > mc.ttl {
		return nil, false
	}
	return mc.models, true
}

// Set updates the cache
func (mc *ModelCache) Set(models []ModelEntry) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.models = models
	mc.cachedAt = time.Now()
}

// ScanModels fetches models from all providers, using cache if valid
func ScanModels(ctx context.Context, endpoints config.EndpointsConfig, cache *ModelCache) []ModelEntry {
	if cached, ok := cache.Get(); ok {
		return cached
	}

	var (
		mu      sync.Mutex
		models  []ModelEntry
		wg      sync.WaitGroup
	)

	for _, provider := range endpoints.Providers {
		wg.Add(1)
		go func(p config.ProviderConfig) {
			defer wg.Done()
			ids, err := FetchModels(ctx, p.ModelsURL)
			if err != nil {
				return // silently skip unavailable providers
			}
			mu.Lock()
			for _, id := range ids {
				models = append(models, ModelEntry{ID: id, Provider: p.Name})
			}
			mu.Unlock()
		}(provider)
	}

	wg.Wait()
	cache.Set(models)
	return models
}

// ModelIDs extracts just the IDs from model entries
func ModelIDs(entries []ModelEntry) []string {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	return ids
}
