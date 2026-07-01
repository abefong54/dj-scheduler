package lead

import (
	"context"

	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type UseCase struct {
	repo repository.LeadRepository
}

func New(repo repository.LeadRepository) *UseCase { return &UseCase{repo: repo} }

// Create persists a lead. Field validation lives in the HTTP handler (codebase
// convention); the usecase only fills the default source.
func (uc *UseCase) Create(ctx context.Context, lead model.Lead) (model.Lead, error) {
	if lead.Source == "" {
		lead.Source = "landing"
	}
	return uc.repo.Create(ctx, lead)
}
