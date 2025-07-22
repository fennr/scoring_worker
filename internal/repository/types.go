package repository

import "time"

// VerificationDataCache представляет запись в таблице verification_data_cache
type VerificationDataCache struct {
	ID        string    `json:"id"`
	DataHash  string    `json:"data_hash"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

// CacheStats представляет статистику использования кэша
type CacheStats struct {
	TotalEntries      int64   `json:"total_entries"`
	CacheHitRate      float64 `json:"cache_hit_rate"`
	StorageSize       int64   `json:"storage_size_bytes"`
	DeduplicationRate float64 `json:"deduplication_rate"`
}
