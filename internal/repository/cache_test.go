package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Интерфейс для pgxpool.Pool
type dbPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Mock для pgxpool.Pool
type mockDBPool struct {
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	execFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockDBPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return nil
}

func (m *mockDBPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

// Mock для pgx.Row
type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	return nil
}

// Mock для pgconn.CommandTag
type mockCommandTag struct {
	rowsAffected int64
}

func (m mockCommandTag) RowsAffected() int64 {
	return m.rowsAffected
}

func (m mockCommandTag) Insert() bool {
	return false
}

func (m mockCommandTag) Update() bool {
	return false
}

func (m mockCommandTag) Delete() bool {
	return false
}

func (m mockCommandTag) Select() bool {
	return false
}

func (m mockCommandTag) String() string {
	return ""
}

// Тестовая версия dataCacheRepository
type testDataCacheRepository struct {
	db     dbPool
	logger *zap.Logger
}

func (r *testDataCacheRepository) ComputeHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (r *testDataCacheRepository) GetDataByHash(ctx context.Context, hash string) (string, error) {
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

func (r *testDataCacheRepository) StoreData(ctx context.Context, data string) (string, error) {
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

func TestComputeHash(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		expectedHash string
	}{
		{
			name:         "simple_json",
			data:         `{"test": "data"}`,
			expectedHash: computeExpectedHash(`{"test": "data"}`),
		},
		{
			name:         "empty_string",
			data:         "",
			expectedHash: computeExpectedHash(""),
		},
		{
			name:         "complex_json",
			data:         `{"company": {"name": "Test Corp", "inn": "1234567890", "data": [1,2,3]}}`,
			expectedHash: computeExpectedHash(`{"company": {"name": "Test Corp", "inn": "1234567890", "data": [1,2,3]}}`),
		},
		{
			name:         "same_data_same_hash",
			data:         `{"test": "data"}`,
			expectedHash: computeExpectedHash(`{"test": "data"}`),
		},
	}

	logger := zaptest.NewLogger(t)
	repo := &testDataCacheRepository{
		db:     nil,
		logger: logger,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := repo.ComputeHash(tt.data)

			if hash != tt.expectedHash {
				t.Errorf("expected hash '%s', but got '%s'", tt.expectedHash, hash)
			}

			// Проверяем, что хэш детерминированный
			hash2 := repo.ComputeHash(tt.data)
			if hash != hash2 {
				t.Errorf("hash should be deterministic, but got different values: '%s' vs '%s'", hash, hash2)
			}
		})
	}
}

func TestGetDataByHash(t *testing.T) {
	tests := []struct {
		name          string
		hash          string
		mockData      string
		mockError     error
		expectedData  string
		expectedError string
	}{
		{
			name:          "successful_get",
			hash:          "abc123",
			mockData:      `{"test": "data"}`,
			mockError:     nil,
			expectedData:  `{"test": "data"}`,
			expectedError: "",
		},
		{
			name:          "data_not_found",
			hash:          "nonexistent",
			mockData:      "",
			mockError:     pgx.ErrNoRows,
			expectedData:  "",
			expectedError: "data not found in cache for hash nonexistent",
		},
		{
			name:          "database_error",
			hash:          "error_hash",
			mockData:      "",
			mockError:     errors.New("database connection failed"),
			expectedData:  "",
			expectedError: "data not found in cache for hash error_hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := &mockDBPool{
				queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return &mockRow{
						scanFunc: func(dest ...any) error {
							if tt.mockError != nil {
								return tt.mockError
							}
							if len(dest) > 0 {
								if strPtr, ok := dest[0].(*string); ok {
									*strPtr = tt.mockData
								}
							}
							return nil
						},
					}
				},
			}

			logger := zaptest.NewLogger(t)
			repo := &testDataCacheRepository{
				db:     mockPool,
				logger: logger,
			}

			data, err := repo.GetDataByHash(context.Background(), tt.hash)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', but got nil", tt.expectedError)
					return
				}
				if !containsError(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', but got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if data != tt.expectedData {
				t.Errorf("expected data '%s', but got '%s'", tt.expectedData, data)
			}
		})
	}
}

func TestStoreData(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		existsInCache bool
		getDataError  error
		execError     error
		expectedError string
		expectedHash  string
	}{
		{
			name:          "successful_store_new_data",
			data:          `{"test": "data"}`,
			existsInCache: false,
			getDataError:  pgx.ErrNoRows,
			expectedHash:  computeExpectedHash(`{"test": "data"}`),
		},
		{
			name:          "data_already_exists",
			data:          `{"test": "data"}`,
			existsInCache: true,
			getDataError:  nil,
			expectedHash:  computeExpectedHash(`{"test": "data"}`),
		},
		{
			name:          "database_exec_error",
			data:          `{"test": "data"}`,
			existsInCache: false,
			getDataError:  pgx.ErrNoRows,
			execError:     errors.New("database error"),
			expectedError: "failed to store data in cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storedHash string
			var storedData string

			mockPool := &mockDBPool{
				queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return &mockRow{
						scanFunc: func(dest ...any) error {
							if tt.existsInCache {
								if len(dest) > 0 {
									if strPtr, ok := dest[0].(*string); ok {
										*strPtr = tt.data
									}
								}
								return nil
							}
							return tt.getDataError
						},
					}
				},
				execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
					if len(args) >= 2 {
						if hash, ok := args[0].(string); ok {
							storedHash = hash
						}
						if data, ok := args[1].(string); ok {
							storedData = data
						}
					}
					if tt.execError != nil {
						return pgconn.CommandTag{}, tt.execError
					}
					return pgconn.CommandTag{}, nil
				},
			}

			logger := zaptest.NewLogger(t)
			repo := &testDataCacheRepository{
				db:     mockPool,
				logger: logger,
			}

			hash, err := repo.StoreData(context.Background(), tt.data)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', but got nil", tt.expectedError)
					return
				}
				if !containsError(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', but got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if hash != tt.expectedHash {
				t.Errorf("expected hash '%s', but got '%s'", tt.expectedHash, hash)
			}

			// Проверяем, что данные были сохранены в БД (если не существовали)
			if !tt.existsInCache && tt.execError == nil {
				if storedHash != tt.expectedHash {
					t.Errorf("expected stored hash '%s', but got '%s'", tt.expectedHash, storedHash)
				}
				if storedData != tt.data {
					t.Errorf("expected stored data '%s', but got '%s'", tt.data, storedData)
				}
			}
		})
	}
}

func TestStoreDataDeduplication(t *testing.T) {
	testData := `{"company": "Test Corp", "inn": "1234567890"}`
	expectedHash := computeExpectedHash(testData)

	var getDataCallCount int
	var execCallCount int

	mockPool := &mockDBPool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			getDataCallCount++
			return &mockRow{
				scanFunc: func(dest ...any) error {
					if getDataCallCount == 1 {
						// Первый вызов - данных нет
						return pgx.ErrNoRows
					}
					// Второй вызов - данные уже есть
					if len(dest) > 0 {
						if strPtr, ok := dest[0].(*string); ok {
							*strPtr = testData
						}
					}
					return nil
				},
			}
		},
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			execCallCount++
			return pgconn.CommandTag{}, nil
		},
	}

	logger := zaptest.NewLogger(t)
	repo := &testDataCacheRepository{
		db:     mockPool,
		logger: logger,
	}

	// Первое сохранение - должно записать в БД
	hash1, err := repo.StoreData(context.Background(), testData)
	if err != nil {
		t.Errorf("unexpected error on first store: %v", err)
		return
	}

	// Второе сохранение тех же данных - должно найти в кэше
	hash2, err := repo.StoreData(context.Background(), testData)
	if err != nil {
		t.Errorf("unexpected error on second store: %v", err)
		return
	}

	// Хэши должны быть одинаковыми
	if hash1 != hash2 {
		t.Errorf("expected same hash for identical data, but got '%s' and '%s'", hash1, hash2)
	}

	if hash1 != expectedHash {
		t.Errorf("expected hash '%s', but got '%s'", expectedHash, hash1)
	}

	// Проверяем, что exec был вызван только один раз (для первого сохранения)
	if execCallCount != 1 {
		t.Errorf("expected exec to be called once, but was called %d times", execCallCount)
	}

	// Проверяем, что getDataByHash был вызван два раза
	if getDataCallCount != 2 {
		t.Errorf("expected getDataByHash to be called twice, but was called %d times", getDataCallCount)
	}
}

func TestHashConsistency(t *testing.T) {
	testCases := []string{
		`{"test": "data"}`,
		`{"company": {"name": "Test", "inn": "123"}}`,
		`[]`,
		`null`,
		`""`,
		`123`,
	}

	logger := zaptest.NewLogger(t)
	repo := &testDataCacheRepository{
		db:     nil,
		logger: logger,
	}

	for _, data := range testCases {
		t.Run("hash_consistency_"+data, func(t *testing.T) {
			hash1 := repo.ComputeHash(data)
			hash2 := repo.ComputeHash(data)

			if hash1 != hash2 {
				t.Errorf("hash should be consistent for data '%s', but got '%s' and '%s'", data, hash1, hash2)
			}

			// Проверяем, что хэш имеет правильную длину (SHA-256 в hex = 64 символа)
			if len(hash1) != 64 {
				t.Errorf("expected hash length 64, but got %d for data '%s'", len(hash1), data)
			}
		})
	}
}

// Вспомогательная функция для вычисления ожидаемого хэша
func computeExpectedHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Вспомогательная функция для проверки содержания ошибки
func containsError(got, want string) bool {
	return len(got) > 0 && len(want) > 0 && (got == want ||
		(len(got) >= len(want) && got[:len(want)] == want))
}
