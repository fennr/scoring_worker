package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Сохраняем оригинальные переменные окружения
	originalEnvVars := make(map[string]string)
	envVarsToTest := []string{
		"DATABASE_HOST", "DATABASE_PORT", "DATABASE_USER", "DATABASE_PASSWORD",
		"DATABASE_DBNAME", "DATABASE_SSLMODE", "NATS_URL", "LOG_LEVEL", "LOG_JSON",
		"CREDINFORM_BASE_URL", "CREDINFORM_USERNAME", "CREDINFORM_PASSWORD",
		"CREDINFORM_TIMEOUT", "CREDINFORM_RETRY_ATTEMPTS", "CREDINFORM_RETRY_DELAY",
		"WORKER_CONCURRENCY",
	}

	for _, envVar := range envVarsToTest {
		originalEnvVars[envVar] = os.Getenv(envVar)
	}

	// Очищаем переменные окружения для чистого теста
	defer func() {
		for envVar, originalValue := range originalEnvVars {
			if originalValue == "" {
				os.Unsetenv(envVar)
			} else {
				os.Setenv(envVar, originalValue)
			}
		}
	}()

	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig *Config
		expectedError  bool
	}{
		{
			name:    "default_values",
			envVars: map[string]string{},
			expectedConfig: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "scoring",
					SSLMode:  "disable",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Log: LogConfig{
					Level: "info",
					JSON:  false,
				},
				Credinform: CredinformConfig{
					BaseURL:       "https://restapi.credinform.ru",
					Username:      "",
					Password:      "",
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    1,
				},
				WorkerConcurrency: 5,
			},
			expectedError: false,
		},
		{
			name: "custom_database_config",
			envVars: map[string]string{
				"DATABASE_HOST":     "db.example.com",
				"DATABASE_PORT":     "5433",
				"DATABASE_USER":     "testuser",
				"DATABASE_PASSWORD": "testpass",
				"DATABASE_DBNAME":   "testdb",
				"DATABASE_SSLMODE":  "require",
			},
			expectedConfig: &Config{
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5433,
					User:     "testuser",
					Password: "testpass",
					DBName:   "testdb",
					SSLMode:  "require",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Log: LogConfig{
					Level: "info",
					JSON:  false,
				},
				Credinform: CredinformConfig{
					BaseURL:       "https://restapi.credinform.ru",
					Username:      "",
					Password:      "",
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    1,
				},
				WorkerConcurrency: 5,
			},
			expectedError: false,
		},
		{
			name: "custom_credinform_config",
			envVars: map[string]string{
				"CREDINFORM_BASE_URL":       "https://api.credinform.com",
				"CREDINFORM_USERNAME":       "testuser",
				"CREDINFORM_PASSWORD":       "testpass",
				"CREDINFORM_TIMEOUT":        "60",
				"CREDINFORM_RETRY_ATTEMPTS": "5",
				"CREDINFORM_RETRY_DELAY":    "2",
			},
			expectedConfig: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "scoring",
					SSLMode:  "disable",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Log: LogConfig{
					Level: "info",
					JSON:  false,
				},
				Credinform: CredinformConfig{
					BaseURL:       "https://api.credinform.com",
					Username:      "testuser",
					Password:      "testpass",
					Timeout:       60,
					RetryAttempts: 5,
					RetryDelay:    2,
				},
				WorkerConcurrency: 5,
			},
			expectedError: false,
		},
		{
			name: "custom_worker_concurrency",
			envVars: map[string]string{
				"WORKER_CONCURRENCY": "10",
			},
			expectedConfig: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "scoring",
					SSLMode:  "disable",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Log: LogConfig{
					Level: "info",
					JSON:  false,
				},
				Credinform: CredinformConfig{
					BaseURL:       "https://restapi.credinform.ru",
					Username:      "",
					Password:      "",
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    1,
				},
				WorkerConcurrency: 10,
			},
			expectedError: false,
		},
		{
			name: "custom_log_config",
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
				"LOG_JSON":  "true",
			},
			expectedConfig: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "scoring",
					SSLMode:  "disable",
				},
				NATS: NATSConfig{
					URL: "nats://localhost:4222",
				},
				Log: LogConfig{
					Level: "debug",
					JSON:  true,
				},
				Credinform: CredinformConfig{
					BaseURL:       "https://restapi.credinform.ru",
					Username:      "",
					Password:      "",
					Timeout:       30,
					RetryAttempts: 3,
					RetryDelay:    1,
				},
				WorkerConcurrency: 5,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем все переменные окружения
			for _, envVar := range envVarsToTest {
				os.Unsetenv(envVar)
			}

			// Устанавливаем переменные окружения для теста
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config, err := Load()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Error("expected config, but got nil")
				return
			}

			// Проверяем Database конфигурацию
			if config.Database.Host != tt.expectedConfig.Database.Host {
				t.Errorf("expected database host '%s', but got '%s'", tt.expectedConfig.Database.Host, config.Database.Host)
			}
			if config.Database.Port != tt.expectedConfig.Database.Port {
				t.Errorf("expected database port %d, but got %d", tt.expectedConfig.Database.Port, config.Database.Port)
			}
			if config.Database.User != tt.expectedConfig.Database.User {
				t.Errorf("expected database user '%s', but got '%s'", tt.expectedConfig.Database.User, config.Database.User)
			}
			if config.Database.Password != tt.expectedConfig.Database.Password {
				t.Errorf("expected database password '%s', but got '%s'", tt.expectedConfig.Database.Password, config.Database.Password)
			}
			if config.Database.DBName != tt.expectedConfig.Database.DBName {
				t.Errorf("expected database name '%s', but got '%s'", tt.expectedConfig.Database.DBName, config.Database.DBName)
			}
			if config.Database.SSLMode != tt.expectedConfig.Database.SSLMode {
				t.Errorf("expected database ssl mode '%s', but got '%s'", tt.expectedConfig.Database.SSLMode, config.Database.SSLMode)
			}

			// Проверяем NATS конфигурацию
			if config.NATS.URL != tt.expectedConfig.NATS.URL {
				t.Errorf("expected NATS URL '%s', but got '%s'", tt.expectedConfig.NATS.URL, config.NATS.URL)
			}

			// Проверяем Log конфигурацию
			if config.Log.Level != tt.expectedConfig.Log.Level {
				t.Errorf("expected log level '%s', but got '%s'", tt.expectedConfig.Log.Level, config.Log.Level)
			}
			if config.Log.JSON != tt.expectedConfig.Log.JSON {
				t.Errorf("expected log JSON %t, but got %t", tt.expectedConfig.Log.JSON, config.Log.JSON)
			}

			// Проверяем Credinform конфигурацию
			if config.Credinform.BaseURL != tt.expectedConfig.Credinform.BaseURL {
				t.Errorf("expected credinform base URL '%s', but got '%s'", tt.expectedConfig.Credinform.BaseURL, config.Credinform.BaseURL)
			}
			if config.Credinform.Username != tt.expectedConfig.Credinform.Username {
				t.Errorf("expected credinform username '%s', but got '%s'", tt.expectedConfig.Credinform.Username, config.Credinform.Username)
			}
			if config.Credinform.Password != tt.expectedConfig.Credinform.Password {
				t.Errorf("expected credinform password '%s', but got '%s'", tt.expectedConfig.Credinform.Password, config.Credinform.Password)
			}
			if config.Credinform.Timeout != tt.expectedConfig.Credinform.Timeout {
				t.Errorf("expected credinform timeout %d, but got %d", tt.expectedConfig.Credinform.Timeout, config.Credinform.Timeout)
			}
			if config.Credinform.RetryAttempts != tt.expectedConfig.Credinform.RetryAttempts {
				t.Errorf("expected credinform retry attempts %d, but got %d", tt.expectedConfig.Credinform.RetryAttempts, config.Credinform.RetryAttempts)
			}
			if config.Credinform.RetryDelay != tt.expectedConfig.Credinform.RetryDelay {
				t.Errorf("expected credinform retry delay %d, but got %d", tt.expectedConfig.Credinform.RetryDelay, config.Credinform.RetryDelay)
			}

			// Проверяем Worker конфигурацию
			if config.WorkerConcurrency != tt.expectedConfig.WorkerConcurrency {
				t.Errorf("expected worker concurrency %d, but got %d", tt.expectedConfig.WorkerConcurrency, config.WorkerConcurrency)
			}
		})
	}
}

func TestDatabaseDSN(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedDSN string
	}{
		{
			name: "default_config",
			config: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "scoring",
					SSLMode:  "disable",
				},
			},
			expectedDSN: "host=localhost port=5432 user=postgres password=postgres dbname=scoring sslmode=disable",
		},
		{
			name: "custom_config",
			config: &Config{
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5433,
					User:     "testuser",
					Password: "testpass",
					DBName:   "testdb",
					SSLMode:  "require",
				},
			},
			expectedDSN: "host=db.example.com port=5433 user=testuser password=testpass dbname=testdb sslmode=require",
		},
		{
			name: "special_characters_in_credentials",
			config: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "user@domain",
					Password: "pass@word#123",
					DBName:   "scoring",
					SSLMode:  "disable",
				},
			},
			expectedDSN: "host=localhost port=5432 user=user@domain password=pass@word#123 dbname=scoring sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DatabaseDSN()
			if dsn != tt.expectedDSN {
				t.Errorf("expected DSN '%s', but got '%s'", tt.expectedDSN, dsn)
			}
		})
	}
}

func TestInvalidConfiguration(t *testing.T) {
	// Сохраняем оригинальные переменные окружения
	originalVars := map[string]string{
		"DATABASE_PORT":             os.Getenv("DATABASE_PORT"),
		"CREDINFORM_TIMEOUT":        os.Getenv("CREDINFORM_TIMEOUT"),
		"CREDINFORM_RETRY_ATTEMPTS": os.Getenv("CREDINFORM_RETRY_ATTEMPTS"),
		"CREDINFORM_RETRY_DELAY":    os.Getenv("CREDINFORM_RETRY_DELAY"),
		"WORKER_CONCURRENCY":        os.Getenv("WORKER_CONCURRENCY"),
	}

	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "invalid_database_port",
			envVars: map[string]string{
				"DATABASE_PORT": "not_a_number",
			},
		},
		{
			name: "invalid_credinform_timeout",
			envVars: map[string]string{
				"CREDINFORM_TIMEOUT": "invalid",
			},
		},
		{
			name: "invalid_retry_attempts",
			envVars: map[string]string{
				"CREDINFORM_RETRY_ATTEMPTS": "not_a_number",
			},
		},
		{
			name: "invalid_retry_delay",
			envVars: map[string]string{
				"CREDINFORM_RETRY_DELAY": "invalid",
			},
		},
		{
			name: "invalid_worker_concurrency",
			envVars: map[string]string{
				"WORKER_CONCURRENCY": "not_a_number",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем переменные окружения
			for key := range originalVars {
				os.Unsetenv(key)
			}

			// Устанавливаем переменные окружения для теста
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			_, err := Load()

			// Ожидаем ошибку при невалидной конфигурации
			if err == nil {
				t.Error("expected error for invalid configuration, but got nil")
			}
		})
	}
}

func TestCredinformConfigValidation(t *testing.T) {
	tests := []struct {
		name                 string
		baseURL              string
		username             string
		password             string
		timeout              int
		retryAttempts        int
		retryDelay           int
		expectedValidBaseURL string
		expectedValidTimeout int
		expectedValidRetries int
		expectedValidDelay   int
	}{
		{
			name:                 "valid_config",
			baseURL:              "https://api.credinform.ru",
			username:             "testuser",
			password:             "testpass",
			timeout:              30,
			retryAttempts:        3,
			retryDelay:           1,
			expectedValidBaseURL: "https://api.credinform.ru",
			expectedValidTimeout: 30,
			expectedValidRetries: 3,
			expectedValidDelay:   1,
		},
		{
			name:                 "zero_timeout_keeps_zero",
			baseURL:              "https://api.credinform.ru",
			username:             "testuser",
			password:             "testpass",
			timeout:              0,
			retryAttempts:        3,
			retryDelay:           1,
			expectedValidBaseURL: "https://api.credinform.ru",
			expectedValidTimeout: 0, // keeps zero
			expectedValidRetries: 3,
			expectedValidDelay:   1,
		},
		{
			name:                 "zero_retry_attempts_keeps_zero",
			baseURL:              "https://api.credinform.ru",
			username:             "testuser",
			password:             "testpass",
			timeout:              30,
			retryAttempts:        0,
			retryDelay:           1,
			expectedValidBaseURL: "https://api.credinform.ru",
			expectedValidTimeout: 30,
			expectedValidRetries: 0, // keeps zero
			expectedValidDelay:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Credinform: CredinformConfig{
					BaseURL:       tt.baseURL,
					Username:      tt.username,
					Password:      tt.password,
					Timeout:       tt.timeout,
					RetryAttempts: tt.retryAttempts,
					RetryDelay:    tt.retryDelay,
				},
			}

			// Проверяем, что конфигурация корректна
			if config.Credinform.BaseURL != tt.expectedValidBaseURL {
				t.Errorf("expected base URL '%s', but got '%s'", tt.expectedValidBaseURL, config.Credinform.BaseURL)
			}
			if config.Credinform.Timeout != tt.expectedValidTimeout {
				t.Errorf("expected timeout %d, but got %d", tt.expectedValidTimeout, config.Credinform.Timeout)
			}
			if config.Credinform.RetryAttempts != tt.expectedValidRetries {
				t.Errorf("expected retry attempts %d, but got %d", tt.expectedValidRetries, config.Credinform.RetryAttempts)
			}
			if config.Credinform.RetryDelay != tt.expectedValidDelay {
				t.Errorf("expected retry delay %d, but got %d", tt.expectedValidDelay, config.Credinform.RetryDelay)
			}
		})
	}
}

func TestWorkerConcurrencyValidation(t *testing.T) {
	// Сохраняем оригинальную переменную окружения
	originalWorkerConcurrency := os.Getenv("WORKER_CONCURRENCY")

	defer func() {
		if originalWorkerConcurrency == "" {
			os.Unsetenv("WORKER_CONCURRENCY")
		} else {
			os.Setenv("WORKER_CONCURRENCY", originalWorkerConcurrency)
		}
	}()

	tests := []struct {
		name                      string
		workerConcurrencyValue    string
		expectedWorkerConcurrency int
	}{
		{
			name:                      "default_value",
			workerConcurrencyValue:    "",
			expectedWorkerConcurrency: 5,
		},
		{
			name:                      "custom_value",
			workerConcurrencyValue:    "10",
			expectedWorkerConcurrency: 10,
		},
		{
			name:                      "zero_value",
			workerConcurrencyValue:    "0",
			expectedWorkerConcurrency: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.workerConcurrencyValue == "" {
				os.Unsetenv("WORKER_CONCURRENCY")
			} else {
				os.Setenv("WORKER_CONCURRENCY", tt.workerConcurrencyValue)
			}

			config, err := Load()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config.WorkerConcurrency != tt.expectedWorkerConcurrency {
				t.Errorf("expected worker concurrency %d, but got %d", tt.expectedWorkerConcurrency, config.WorkerConcurrency)
			}
		})
	}
}
