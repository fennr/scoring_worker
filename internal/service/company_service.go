package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"scoring_worker/internal/credinform"
	"scoring_worker/internal/repository"

	"go.uber.org/zap"
)

type CompanyService struct {
	credinformClient *credinform.Client
	repo             repository.VerificationRepository
	logger           *zap.Logger
}

func NewCompanyService(client *credinform.Client, repo repository.VerificationRepository, logger *zap.Logger) *CompanyService {
	return &CompanyService{
		credinformClient: client,
		repo:             repo,
		logger:           logger,
	}
}

func (s *CompanyService) ProcessVerification(ctx context.Context, verificationID, inn string, requestedTypes []string) error {
	s.logger.Info("Starting verification processing",
		zap.String("verification_id", verificationID),
		zap.String("inn", inn),
		zap.Strings("requested_types", requestedTypes))

	// Шаг 1: Поиск компании по ИНН
	companyData, err := s.credinformClient.SearchCompany(ctx, inn)
	if err != nil {
		s.logger.Error("Failed to search company",
			zap.Error(err),
			zap.String("inn", inn))

		// Если компания не найдена, обновляем статус на CompanyNotFound
		if err.Error() == "company not found" {
			if updateErr := s.repo.UpdateStatus(ctx, verificationID, "COMPANY_NOT_FOUND"); updateErr != nil {
				s.logger.Error("Failed to update status to COMPANY_NOT_FOUND", zap.Error(updateErr))
			}
			return fmt.Errorf("company not found: %s", inn)
		}

		// Для других ошибок обновляем статус на ERROR
		if updateErr := s.repo.UpdateStatus(ctx, verificationID, "ERROR"); updateErr != nil {
			s.logger.Error("Failed to update status to ERROR", zap.Error(updateErr))
		}
		return fmt.Errorf("failed to search company: %w", err)
	}

	// Шаг 2: Сохраняем company_id в БД
	if err := s.repo.UpdateCompanyID(ctx, verificationID, companyData.CompanyID); err != nil {
		s.logger.Error("Failed to update company_id",
			zap.Error(err),
			zap.String("verification_id", verificationID),
			zap.String("company_id", companyData.CompanyID))
		return fmt.Errorf("failed to update company_id: %w", err)
	}

	// Шаг 3: Обновляем статус на PROCESSING
	if err := s.repo.UpdateStatus(ctx, verificationID, "PROCESSING"); err != nil {
		s.logger.Error("Failed to update status to PROCESSING",
			zap.Error(err),
			zap.String("verification_id", verificationID))
		return fmt.Errorf("failed to update status to PROCESSING: %w", err)
	}

	// Шаг 4: Обрабатываем запрошенные типы данных
	for _, dataType := range requestedTypes {
		var dataForDB interface{}
		var err error

		switch strings.ToLower(dataType) {
		case "basic_information":
			dataForDB, err = s.credinformClient.GetBasicInformation(ctx, companyData.CompanyID, credinform.BasicInformationParams{})
		case "activities":
			dataForDB, err = s.credinformClient.GetActivities(ctx, companyData.CompanyID, credinform.ActivitiesParams{})
		case "addresses_by_credinform":
			dataForDB, err = s.credinformClient.GetAddressesByCredinform(ctx, companyData.CompanyID, credinform.AddressesByCredinformParams{})
		case "addresses_by_unified_state_register":
			dataForDB, err = s.credinformClient.GetAddressesByUnifiedStateRegister(ctx, companyData.CompanyID, credinform.AddressesByUnifiedStateRegisterParams{})
		case "affiliated_companies":
			params := credinform.AffiliatedCompaniesParams{
				AffiliationTypes: []string{
					"ByManagementOrShareholdersNaturalPersons",
					"ByLiquidatorOrBankruptcyAdministrator",
					"UnderAdministrationOfTheCompany",
					"ByManagingLegalPersons",
					"ByShareholdersLegalPersons",
				},
			}
			dataForDB, err = s.credinformClient.GetAffiliatedCompanies(ctx, companyData.CompanyID, params)
		case "arbitrage_statistics":
			params := credinform.ArbitrageStatisticsParams{
				ArbitrageSideCommonType: []string{
					"Claimant",
					"Defendant",
					"ThirdPartiesAndOthers",
				},
			}
			params.LastCaseChangeDateRange.From = "2025-01-01T00:00:00"
			dataForDB, err = s.credinformClient.GetArbitrageStatistics(ctx, companyData.CompanyID, params)
		default:
			s.logger.Warn("Unknown data type requested", zap.String("type", dataType))
			continue
		}

		if err != nil {
			s.logger.Error("Failed to get company data",
				zap.Error(err),
				zap.String("type", dataType),
				zap.String("company_id", companyData.CompanyID))

			errorData := map[string]interface{}{
				"error":        err.Error(),
				"type":         dataType,
				"company_id":   companyData.CompanyID,
				"processed_at": time.Now().Format(time.RFC3339),
			}

			dataJSON, _ := json.Marshal(errorData)
			if err := s.repo.AddData(ctx, verificationID, dataType, string(dataJSON)); err != nil {
				s.logger.Error("Failed to add error data to repository", zap.Error(err))
			}
			continue
		}

		resultData := map[string]interface{}{
			"data":         dataForDB,
			"type":         dataType,
			"company_id":   companyData.CompanyID,
			"processed_at": time.Now().Format(time.RFC3339),
			"status":       "completed",
		}

		dataJSON, err := json.Marshal(resultData)
		if err != nil {
			s.logger.Error("Failed to marshal result data", zap.Error(err))
			continue
		}

		if err := s.repo.AddData(ctx, verificationID, dataType, string(dataJSON)); err != nil {
			s.logger.Error("Failed to add verification data",
				zap.Error(err),
				zap.String("verification_id", verificationID),
				zap.String("data_type", dataType))
		}
	}

	// Шаг 5: Обновляем статус на COMPLETED
	if err := s.repo.UpdateStatus(ctx, verificationID, "COMPLETED"); err != nil {
		s.logger.Error("Failed to update status to COMPLETED",
			zap.Error(err),
			zap.String("verification_id", verificationID))
		return fmt.Errorf("failed to update status: %w", err)
	}

	s.logger.Info("Verification processing completed",
		zap.String("verification_id", verificationID),
		zap.String("company_id", companyData.CompanyID))
	return nil
}
