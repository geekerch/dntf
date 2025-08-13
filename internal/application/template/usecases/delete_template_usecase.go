package usecases

import (
	"context"
	"fmt"

	"notification/internal/domain/template"
)

// DeleteTemplateUseCase handles deleting templates.
type DeleteTemplateUseCase struct {
	templateRepo template.TemplateRepository
}

// NewDeleteTemplateUseCase creates a new DeleteTemplateUseCase.
func NewDeleteTemplateUseCase(templateRepo template.TemplateRepository) *DeleteTemplateUseCase {
	return &DeleteTemplateUseCase{
		templateRepo: templateRepo,
	}
}

// Execute deletes a template.
func (uc *DeleteTemplateUseCase) Execute(ctx context.Context, id string) error {
	// Validate input
	if id == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	// Create template ID
	templateID, err := template.NewTemplateIDFromString(id)
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	// Check if template exists
	exists, err := uc.templateRepo.Exists(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to check template existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("template with ID '%s' not found", id)
	}

	// Delete template
	if err := uc.templateRepo.Delete(ctx, templateID); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}