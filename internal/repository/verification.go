package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type VerificationRepository interface {
	Create(ctx context.Context, id string, inn string, requestedTypes []string, authorEmail string) error
	UpdateStatus(ctx context.Context, id string, status string) error
	UpdateCompanyID(ctx context.Context, id string, companyID string) error
	AddData(ctx context.Context, verificationID string, dataType string, data string) error
	GetByID(ctx context.Context, id string) (*Verification, error)
	AcquireStaleVerification(ctx context.Context) (*Verification, error)
}

type Verification struct {
	ID                 string    `json:"id"`
	Inn                string    `json:"inn"`
	Status             string    `json:"status"`
	AuthorEmail        string    `json:"author_email"`
	CompanyID          string    `json:"company_id"`
	RequestedDataTypes []string  `json:"requested_data_types"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type verificationRepository struct {
	db        *pgxpool.Pool
	cacheRepo DataCacheRepository
	logger    *zap.Logger
}

func NewVerificationRepository(db *pgxpool.Pool, cacheRepo DataCacheRepository, logger *zap.Logger) VerificationRepository {
	return &verificationRepository{
		db:        db,
		cacheRepo: cacheRepo,
		logger:    logger,
	}
}

func (r *verificationRepository) Create(ctx context.Context, id string, inn string, requestedTypes []string, authorEmail string) error {
	now := time.Now().Format(time.RFC3339)

	query := `
		INSERT INTO verifications (id, inn, status, author_email, requested_data_types, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query, id, inn, "IN_PROCESS", authorEmail, requestedTypes, now, now)
	if err != nil {
		r.logger.Error("failed to create verification", zap.Error(err), zap.String("id", id), zap.String("inn", inn))
		return fmt.Errorf("failed to create verification: %w", err)
	}

	r.logger.Info("verification created", zap.String("id", id), zap.String("inn", inn))
	return nil
}

func (r *verificationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE verifications SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		r.logger.Error("failed to update verification status", zap.Error(err), zap.String("id", id), zap.String("status", status))
		return fmt.Errorf("failed to update verification status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("verification not found: %s", id)
	}
	r.logger.Info("verification status updated", zap.String("id", id), zap.String("status", status))
	return nil
}

func (r *verificationRepository) UpdateCompanyID(ctx context.Context, id string, companyID string) error {
	query := `UPDATE verifications SET company_id = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(ctx, query, companyID, id)
	if err != nil {
		r.logger.Error("failed to update verification company_id", zap.Error(err), zap.String("id", id), zap.String("company_id", companyID))
		return fmt.Errorf("failed to update verification company_id: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("verification not found: %s", id)
	}
	r.logger.Info("verification company_id updated", zap.String("id", id), zap.String("company_id", companyID))
	return nil
}

// AddData добавляет данные проверки с использованием системы кэширования
func (r *verificationRepository) AddData(ctx context.Context, verificationID string, dataType string, data string) error {
	dataHash, err := r.cacheRepo.StoreData(ctx, data)
	if err != nil {
		r.logger.Error("failed to store data in cache", zap.Error(err), zap.String("verification_id", verificationID))
		return fmt.Errorf("failed to store data in cache: %w", err)
	}

	query := `
		INSERT INTO verification_data (verification_id, data_type, data_hash, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (verification_id, data_type) DO UPDATE SET
			data_hash = EXCLUDED.data_hash,
			created_at = NOW()
	`

	_, err = r.db.Exec(ctx, query, verificationID, dataType, dataHash)
	if err != nil {
		r.logger.Error("failed to add verification data", zap.Error(err), zap.String("verification_id", verificationID))
		return fmt.Errorf("failed to add verification data: %w", err)
	}

	r.logger.Info("verification data added", zap.String("verification_id", verificationID), zap.String("data_type", dataType))
	return nil
}

func (r *verificationRepository) GetByID(ctx context.Context, id string) (*Verification, error) {
	query := `
		SELECT id, inn, status, author_email, company_id, requested_data_types, created_at, updated_at
		FROM verifications
		WHERE id = $1
	`

	var verification Verification
	err := r.db.QueryRow(ctx, query, id).
		Scan(&verification.ID, &verification.Inn, &verification.Status, &verification.AuthorEmail, &verification.CompanyID, &verification.RequestedDataTypes, &verification.CreatedAt, &verification.UpdatedAt)
	if err != nil {
		r.logger.Error("failed to get verification", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	return &verification, nil
}

func (r *verificationRepository) AcquireStaleVerification(ctx context.Context) (*Verification, error) {
	query := `
		UPDATE verifications
		SET status = 'PROCESSING', updated_at = NOW()
		WHERE id = (
			SELECT id
			FROM verifications
			WHERE status = 'IN_PROCESS' OR status = 'PROCESSING'
			ORDER BY updated_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, inn, status, author_email, company_id, requested_data_types, created_at, updated_at
	`
	var v Verification
	err := r.db.QueryRow(ctx, query).Scan(
		&v.ID, &v.Inn, &v.Status, &v.AuthorEmail, &v.CompanyID,
		&v.RequestedDataTypes, &v.CreatedAt, &v.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" { // pgx specific error
			return nil, nil // No stale verification found, not an error
		}
		r.logger.Error("failed to acquire stale verification", zap.Error(err))
		return nil, fmt.Errorf("failed to acquire stale verification: %w", err)
	}

	return &v, nil
}
