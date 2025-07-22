package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DataCacheRepository interface {
	GetDataByHash(ctx context.Context, hash string) (string, error)
	StoreData(ctx context.Context, data string) (string, error)
	ComputeHash(data string) string
}

type dataCacheRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewDataCacheRepository(db *pgxpool.Pool, logger *zap.Logger) DataCacheRepository {
	return &dataCacheRepository{
		db:     db,
		logger: logger,
	}
}

// ComputeHash вычисляет SHA-256 хэш от JSON строки
func (r *dataCacheRepository) ComputeHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetDataByHash получает данные из кэша по хэшу
func (r *dataCacheRepository) GetDataByHash(ctx context.Context, hash string) (string, error) {
	query := `SELECT data FROM verification_data_cache WHERE data_hash = $1`

	var data string
	err := r.db.QueryRow(ctx, query, hash).Scan(&data)
	if err != nil {
		r.logger.Debug("data not found in cache", zap.String("hash", hash), zap.Error(err))
		return "", fmt.Errorf("data not found in cache for hash %s: %w", hash, err)
	}

	r.logger.Debug("data retrieved from cache", zap.String("hash", hash))
	return data, nil
}

// StoreData сохраняет данные в кэш и возвращает их хэш
func (r *dataCacheRepository) StoreData(ctx context.Context, data string) (string, error) {
	hash := r.ComputeHash(data)

	_, err := r.GetDataByHash(ctx, hash)
	if err == nil {
		r.logger.Debug("data already exists in cache", zap.String("hash", hash))
		return hash, nil
	}

	query := `
		INSERT INTO verification_data_cache (data_hash, data, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (data_hash) DO NOTHING
	`

	_, err = r.db.Exec(ctx, query, hash, data)
	if err != nil {
		r.logger.Error("failed to store data in cache", zap.Error(err), zap.String("hash", hash))
		return "", fmt.Errorf("failed to store data in cache: %w", err)
	}

	r.logger.Info("data stored in cache", zap.String("hash", hash))
	return hash, nil
}
