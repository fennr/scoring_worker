package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
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

	companyData, err := s.searchCompany(ctx, verificationID, inn)
	if err != nil {
		return err
	}

	if err := s.prepareVerification(ctx, verificationID, companyData.CompanyID); err != nil {
		return err
	}

	if err := s.updateVerificationStatus(ctx, verificationID, "PROCESSING"); err != nil {
		return err
	}

	s.processDataTypes(ctx, verificationID, companyData.CompanyID, requestedTypes)

	if err := s.updateVerificationStatus(ctx, verificationID, "COMPLETED"); err != nil {
		return err
	}

	s.logger.Info("Verification processing completed",
		zap.String("verification_id", verificationID),
		zap.String("company_id", companyData.CompanyID))

	return nil
}

func (s *CompanyService) searchCompany(ctx context.Context, verificationID, inn string) (*credinform.CompanyData, error) {
	companyData, err := s.credinformClient.SearchCompany(ctx, inn)
	if err != nil {
		s.logger.Error("Failed to search company", zap.Error(err), zap.String("inn", inn))
		status := "ERROR"
		if err.Error() == "company not found" {
			status = "COMPANY_NOT_FOUND"
		}
		_ = s.updateVerificationStatus(ctx, verificationID, status)
		return nil, fmt.Errorf("failed to search company (inn: %s): %w", inn, err)
	}
	return companyData, nil
}

func (s *CompanyService) prepareVerification(ctx context.Context, verificationID, companyID string) error {
	if err := s.repo.UpdateCompanyID(ctx, verificationID, companyID); err != nil {
		s.logger.Error("Failed to update company_id",
			zap.Error(err),
			zap.String("verification_id", verificationID),
			zap.String("company_id", companyID))
		return fmt.Errorf("failed to update company_id: %w", err)
	}

	return nil
}

func (s *CompanyService) processDataTypes(ctx context.Context, verificationID, companyID string, requestedTypes []string) {
	var wg sync.WaitGroup
	for _, dataType := range requestedTypes {
		wg.Add(1)
		go s.fetchAndSaveData(ctx, &wg, verificationID, companyID, dataType)
	}
	wg.Wait()
}

func (s *CompanyService) fetchAndSaveData(ctx context.Context, wg *sync.WaitGroup, verificationID, companyID, dataType string) {
	defer wg.Done()

	var dataForDB interface{}
	var err error

	switch strings.ToLower(dataType) {
	case "basic_information":
		dataForDB, err = s.credinformClient.GetBasicInformation(ctx, companyID, credinform.BasicInformationParams{})
	case "activities":
		dataForDB, err = s.credinformClient.GetActivities(ctx, companyID, credinform.ActivitiesParams{})
	case "addresses_by_credinform":
		dataForDB, err = s.credinformClient.GetAddressesByCredinform(ctx, companyID, credinform.AddressesByCredinformParams{})
	case "addresses_by_unified_state_register":
		dataForDB, err = s.credinformClient.GetAddressesByUnifiedStateRegister(ctx, companyID, credinform.AddressesByUnifiedStateRegisterParams{})
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
		dataForDB, err = s.credinformClient.GetAffiliatedCompanies(ctx, companyID, params)
	case "arbitrage_statistics":
		params := credinform.ArbitrageStatisticsParams{
			ArbitrageSideCommonType: []string{
				"Claimant",
				"Defendant",
				"ThirdPartiesAndOthers",
			},
		}
		params.LastCaseChangeDateRange.From = "2025-01-01T00:00:00"
		dataForDB, err = s.credinformClient.GetArbitrageStatistics(ctx, companyID, params)
	default:
		s.logger.Warn("Unknown data type requested", zap.String("type", dataType))
		return
	}

	if err != nil {
		s.logger.Error("Failed to get company data",
			zap.Error(err),
			zap.String("type", dataType),
			zap.String("company_id", companyID))
		s.saveErrorData(ctx, verificationID, companyID, dataType, err)
		return
	}

	s.saveSuccessData(ctx, verificationID, companyID, dataType, dataForDB)
}

func (s *CompanyService) saveErrorData(ctx context.Context, verificationID, companyID, dataType string, err error) {
	errorData := map[string]interface{}{
		"error":        err.Error(),
		"type":         dataType,
		"company_id":   companyID,
		"processed_at": time.Now().Format(time.RFC3339),
	}
	dataJSON, _ := json.Marshal(errorData)
	if dbErr := s.repo.AddData(ctx, verificationID, dataType, string(dataJSON)); dbErr != nil {
		s.logger.Error("Failed to add error data to repository", zap.Error(dbErr))
	}
}

func (s *CompanyService) saveSuccessData(ctx context.Context, verificationID, companyID, dataType string, dataForDB interface{}) {
	resultData := map[string]interface{}{
		"data":         dataForDB,
		"type":         dataType,
		"company_id":   companyID,
		"processed_at": time.Now().Format(time.RFC3339),
		"status":       "completed",
	}

	dataJSON, err := json.Marshal(resultData)
	if err != nil {
		s.logger.Error("Failed to marshal result data", zap.Error(err))
		s.saveErrorData(ctx, verificationID, companyID, dataType, err)
		return
	}

	if err := s.repo.AddData(ctx, verificationID, dataType, string(dataJSON)); err != nil {
		s.logger.Error("Failed to add verification data",
			zap.Error(err),
			zap.String("verification_id", verificationID),
			zap.String("data_type", dataType))
	}
}

func (s *CompanyService) updateVerificationStatus(ctx context.Context, verificationID, status string) error {
	if err := s.repo.UpdateStatus(ctx, verificationID, status); err != nil {
		s.logger.Error("Failed to update status",
			zap.Error(err),
			zap.String("verification_id", verificationID),
			zap.String("status", status))
		return fmt.Errorf("failed to update status to %s: %w", status, err)
	}
	return nil
}
