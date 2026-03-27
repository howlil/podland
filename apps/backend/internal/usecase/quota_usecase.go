package usecase

import (
	"context"

	"github.com/podland/backend/internal/entity"
	"github.com/podland/backend/internal/repository"
	pkgerrors "github.com/podland/backend/pkg/errors"
)

// QuotaUsecase defines quota business logic
type QuotaUsecase struct {
	quotaRepo repository.QuotaRepository
}

// NewQuotaUsecase creates a new quota usecase with dependencies
func NewQuotaUsecase(quotaRepo repository.QuotaRepository) *QuotaUsecase {
	return &QuotaUsecase{
		quotaRepo: quotaRepo,
	}
}

// GetQuota gets the quota for a user
func (uc *QuotaUsecase) GetQuota(ctx context.Context, userID string) (*entity.Quota, error) {
	quota, err := uc.quotaRepo.GetQuota(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to get quota")
	}

	return quota, nil
}

// GetAllTiers gets all available tiers
func (uc *QuotaUsecase) GetAllTiers(ctx context.Context) ([]*entity.Tier, error) {
	tiers, err := uc.quotaRepo.GetAllTiers(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to get tiers")
	}

	return tiers, nil
}

// GetTier gets a specific tier by name
func (uc *QuotaUsecase) GetTier(ctx context.Context, name string) (*entity.Tier, error) {
	tier, err := uc.quotaRepo.GetTier(ctx, name)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to get tier")
	}

	return tier, nil
}
