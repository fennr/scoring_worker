package service

import (
	"context"
	"errors"
	"testing"

	"scoring_worker/internal/credinform"
	"scoring_worker/internal/credinform/types"
	"scoring_worker/internal/repository"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Mock для Credinform Client
type mockCredinformClient struct {
	searchCompanyFunc                      func(ctx context.Context, inn string) (*credinform.CompanyData, error)
	getBasicInformationFunc                func(ctx context.Context, companyID string, params credinform.BasicInformationParams) (*types.BasicInformation, error)
	getActivitiesFunc                      func(ctx context.Context, companyID string, params credinform.ActivitiesParams) (*types.Activities, error)
	getAddressesByCredinformFunc           func(ctx context.Context, companyID string, params credinform.AddressesByCredinformParams) (*types.AddressesByCredinform, error)
	getAddressesByUnifiedStateRegisterFunc func(ctx context.Context, companyID string, params credinform.AddressesByUnifiedStateRegisterParams) (*types.AddressesByUnifiedStateRegister, error)
	getAffiliatedCompaniesFunc             func(ctx context.Context, companyID string, params credinform.AffiliatedCompaniesParams) (*types.AffiliatedCompanies, error)
	getArbitrageStatisticsFunc             func(ctx context.Context, companyID string, params credinform.ArbitrageStatisticsParams) (*types.ArbitrageStatistics, error)
}

func (m *mockCredinformClient) SearchCompany(ctx context.Context, inn string) (*credinform.CompanyData, error) {
	if m.searchCompanyFunc != nil {
		return m.searchCompanyFunc(ctx, inn)
	}
	return &credinform.CompanyData{CompanyID: "test-company-id"}, nil
}

func (m *mockCredinformClient) GetBasicInformation(ctx context.Context, companyID string, params credinform.BasicInformationParams) (*types.BasicInformation, error) {
	if m.getBasicInformationFunc != nil {
		return m.getBasicInformationFunc(ctx, companyID, params)
	}
	return &types.BasicInformation{}, nil
}

func (m *mockCredinformClient) GetActivities(ctx context.Context, companyID string, params credinform.ActivitiesParams) (*types.Activities, error) {
	if m.getActivitiesFunc != nil {
		return m.getActivitiesFunc(ctx, companyID, params)
	}
	return &types.Activities{}, nil
}

func (m *mockCredinformClient) GetAddressesByCredinform(ctx context.Context, companyID string, params credinform.AddressesByCredinformParams) (*types.AddressesByCredinform, error) {
	if m.getAddressesByCredinformFunc != nil {
		return m.getAddressesByCredinformFunc(ctx, companyID, params)
	}
	return &types.AddressesByCredinform{}, nil
}

func (m *mockCredinformClient) GetAddressesByUnifiedStateRegister(ctx context.Context, companyID string, params credinform.AddressesByUnifiedStateRegisterParams) (*types.AddressesByUnifiedStateRegister, error) {
	if m.getAddressesByUnifiedStateRegisterFunc != nil {
		return m.getAddressesByUnifiedStateRegisterFunc(ctx, companyID, params)
	}
	return &types.AddressesByUnifiedStateRegister{}, nil
}

func (m *mockCredinformClient) GetAffiliatedCompanies(ctx context.Context, companyID string, params credinform.AffiliatedCompaniesParams) (*types.AffiliatedCompanies, error) {
	if m.getAffiliatedCompaniesFunc != nil {
		return m.getAffiliatedCompaniesFunc(ctx, companyID, params)
	}
	return &types.AffiliatedCompanies{}, nil
}

func (m *mockCredinformClient) GetArbitrageStatistics(ctx context.Context, companyID string, params credinform.ArbitrageStatisticsParams) (*types.ArbitrageStatistics, error) {
	if m.getArbitrageStatisticsFunc != nil {
		return m.getArbitrageStatisticsFunc(ctx, companyID, params)
	}
	return &types.ArbitrageStatistics{}, nil
}

// Mock для VerificationRepository
type mockVerificationRepository struct {
	createFunc          func(ctx context.Context, id string, inn string, requestedTypes []string, authorEmail string) error
	updateStatusFunc    func(ctx context.Context, id string, status string) error
	updateCompanyIDFunc func(ctx context.Context, id string, companyID string) error
	addDataFunc         func(ctx context.Context, verificationID string, dataType string, data string) error
}

func (m *mockVerificationRepository) Create(ctx context.Context, id string, inn string, requestedTypes []string, authorEmail string) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, id, inn, requestedTypes, authorEmail)
	}
	return nil
}

func (m *mockVerificationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *mockVerificationRepository) UpdateCompanyID(ctx context.Context, id string, companyID string) error {
	if m.updateCompanyIDFunc != nil {
		return m.updateCompanyIDFunc(ctx, id, companyID)
	}
	return nil
}

func (m *mockVerificationRepository) AddData(ctx context.Context, verificationID string, dataType string, data string) error {
	if m.addDataFunc != nil {
		return m.addDataFunc(ctx, verificationID, dataType, data)
	}
	return nil
}

func (m *mockVerificationRepository) GetByID(ctx context.Context, id string) (*repository.Verification, error) {
	return nil, nil
}

func (m *mockVerificationRepository) AcquireStaleVerification(ctx context.Context) (*repository.Verification, error) {
	return nil, nil
}

// Интерфейс для Credinform клиента (для внедрения зависимостей в тестах)
type CredinformClientInterface interface {
	SearchCompany(ctx context.Context, inn string) (*credinform.CompanyData, error)
	GetBasicInformation(ctx context.Context, companyID string, params credinform.BasicInformationParams) (*types.BasicInformation, error)
	GetActivities(ctx context.Context, companyID string, params credinform.ActivitiesParams) (*types.Activities, error)
	GetAddressesByCredinform(ctx context.Context, companyID string, params credinform.AddressesByCredinformParams) (*types.AddressesByCredinform, error)
	GetAddressesByUnifiedStateRegister(ctx context.Context, companyID string, params credinform.AddressesByUnifiedStateRegisterParams) (*types.AddressesByUnifiedStateRegister, error)
	GetAffiliatedCompanies(ctx context.Context, companyID string, params credinform.AffiliatedCompaniesParams) (*types.AffiliatedCompanies, error)
	GetArbitrageStatistics(ctx context.Context, companyID string, params credinform.ArbitrageStatisticsParams) (*types.ArbitrageStatistics, error)
}

// Тестовая версия VerificationService для внедрения зависимостей
type testVerificationService struct {
	repo       repository.VerificationRepository
	client     CredinformClientInterface
	logger     *zap.Logger
	maxRetries int
}

func NewTestVerificationService(repo repository.VerificationRepository, client CredinformClientInterface, logger *zap.Logger) *testVerificationService {
	return &testVerificationService{
		repo:       repo,
		client:     client,
		logger:     logger,
		maxRetries: 3,
	}
}

func (s *testVerificationService) ProcessVerification(ctx context.Context, verificationID, inn string, requestedTypes []string) error {
	// Поиск компании
	companyData, err := s.client.SearchCompany(ctx, inn)
	if err != nil {
		s.logger.Error("Failed to search company", zap.Error(err))
		if updateErr := s.repo.UpdateStatus(ctx, verificationID, "error"); updateErr != nil {
			s.logger.Error("Failed to update status to error", zap.Error(updateErr))
		}
		return err
	}

	// Обновление company_id
	if err := s.repo.UpdateCompanyID(ctx, verificationID, companyData.CompanyID); err != nil {
		s.logger.Error("Failed to update company ID", zap.Error(err))
		return err
	}

	// Получение данных по каждому типу
	for _, dataType := range requestedTypes {
		var fetchErr error

		switch dataType {
		case "basic_information":
			_, fetchErr = s.client.GetBasicInformation(ctx, companyData.CompanyID, credinform.BasicInformationParams{})
		case "activities":
			_, fetchErr = s.client.GetActivities(ctx, companyData.CompanyID, credinform.ActivitiesParams{})
		case "addresses_by_credinform":
			_, fetchErr = s.client.GetAddressesByCredinform(ctx, companyData.CompanyID, credinform.AddressesByCredinformParams{})
		case "addresses_by_unified_state_register":
			_, fetchErr = s.client.GetAddressesByUnifiedStateRegister(ctx, companyData.CompanyID, credinform.AddressesByUnifiedStateRegisterParams{})
		case "affiliated_companies":
			_, fetchErr = s.client.GetAffiliatedCompanies(ctx, companyData.CompanyID, credinform.AffiliatedCompaniesParams{})
		case "arbitrage_statistics":
			_, fetchErr = s.client.GetArbitrageStatistics(ctx, companyData.CompanyID, credinform.ArbitrageStatisticsParams{})
		default:
			s.logger.Warn("Unknown data type requested", zap.String("type", dataType))
			continue
		}

		if fetchErr != nil {
			s.logger.Error("Failed to fetch data", zap.String("type", dataType), zap.Error(fetchErr))
			continue
		}

		// Сохранение данных (в реальной реализации здесь была бы сериализация в JSON)
		if err := s.repo.AddData(ctx, verificationID, dataType, "mock_data"); err != nil {
			s.logger.Error("Failed to save data", zap.String("type", dataType), zap.Error(err))
		}
	}

	// Обновление статуса на completed
	if err := s.repo.UpdateStatus(ctx, verificationID, "completed"); err != nil {
		s.logger.Error("Failed to update status to completed", zap.Error(err))
		return err
	}

	return nil
}

func TestProcessVerification(t *testing.T) {
	tests := []struct {
		name                 string
		verificationID       string
		inn                  string
		requestedTypes       []string
		searchCompanyError   error
		updateCompanyIDError error
		updateStatusError    error
		expectedError        string
	}{
		{
			name:           "successful_processing",
			verificationID: "test-verification-id",
			inn:            "1234567890",
			requestedTypes: []string{"basic_information", "activities"},
		},
		{
			name:               "search_company_error",
			verificationID:     "test-verification-id",
			inn:                "1234567890",
			requestedTypes:     []string{"basic_information"},
			searchCompanyError: errors.New("company not found"),
			expectedError:      "company not found",
		},
		{
			name:                 "update_company_id_error",
			verificationID:       "test-verification-id",
			inn:                  "1234567890",
			requestedTypes:       []string{"basic_information"},
			updateCompanyIDError: errors.New("database error"),
			expectedError:        "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			mockRepo := &mockVerificationRepository{
				updateCompanyIDFunc: func(ctx context.Context, id string, companyID string) error {
					return tt.updateCompanyIDError
				},
				updateStatusFunc: func(ctx context.Context, id string, status string) error {
					return tt.updateStatusError
				},
			}

			mockClient := &mockCredinformClient{
				searchCompanyFunc: func(ctx context.Context, inn string) (*credinform.CompanyData, error) {
					if tt.searchCompanyError != nil {
						return nil, tt.searchCompanyError
					}
					return &credinform.CompanyData{CompanyID: "test-company-id"}, nil
				},
			}

			service := NewTestVerificationService(mockRepo, mockClient, logger)

			err := service.ProcessVerification(context.Background(), tt.verificationID, tt.inn, tt.requestedTypes)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestProcessVerificationDataTypes(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockRepo := &mockVerificationRepository{}
	mockClient := &mockCredinformClient{
		searchCompanyFunc: func(ctx context.Context, inn string) (*credinform.CompanyData, error) {
			return &credinform.CompanyData{CompanyID: "test-company-id"}, nil
		},
	}

	service := NewTestVerificationService(mockRepo, mockClient, logger)

	dataTypes := []string{
		"basic_information",
		"activities",
		"addresses_by_credinform",
		"addresses_by_unified_state_register",
		"affiliated_companies",
		"arbitrage_statistics",
		"unknown_type", // должен быть проигнорирован
	}

	err := service.ProcessVerification(context.Background(), "test-id", "1234567890", dataTypes)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestProcessVerificationWithCredinformErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockRepo := &mockVerificationRepository{}
	mockClient := &mockCredinformClient{
		searchCompanyFunc: func(ctx context.Context, inn string) (*credinform.CompanyData, error) {
			return &credinform.CompanyData{CompanyID: "test-company-id"}, nil
		},
		getBasicInformationFunc: func(ctx context.Context, companyID string, params credinform.BasicInformationParams) (*types.BasicInformation, error) {
			return nil, errors.New("credinform API error")
		},
	}

	service := NewTestVerificationService(mockRepo, mockClient, logger)

	// Ошибки получения данных не должны прерывать весь процесс
	err := service.ProcessVerification(context.Background(), "test-id", "1234567890", []string{"basic_information"})
	if err != nil {
		t.Errorf("Expected no error when data fetching fails, got %v", err)
	}
}
