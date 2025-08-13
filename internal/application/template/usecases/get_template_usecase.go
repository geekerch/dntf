package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/template/dtos"
	"notification/internal/domain/template"
)

// GetTemplateUseCase handles getting a single template.
type GetTemplateUseCase struct {
	templateRepo template.TemplateRepository
}

// NewGetTemplateUseCase creates a new GetTemplateUseCase.
func NewGetTemplateUseCase(templateRepo template.TemplateRepository) *GetTemplateUseCase {
	return &GetTemplateUseCase{
		templateRepo: templateRepo,
	}
}

// Execute gets a template by ID.
func (uc *GetTemplateUseCase) Execute(ctx context.Context, id string) (*dtos.TemplateResponse, error) {
	// Validate input
	if id == "" {
		return nil, fmt.Errorf("template ID cannot be empty")
	}

	// Create template ID
	templateID, err := template.NewTemplateIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	// Find template
	templateEntity, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Convert to response
	return dtos.ToTemplateResponse(templateEntity), nil
}