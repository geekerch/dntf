package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/template/dtos"
	"notification/internal/domain/template"
)

// UpdateTemplateUseCase handles updating templates.
type UpdateTemplateUseCase struct {
	templateRepo template.TemplateRepository
}

// NewUpdateTemplateUseCase creates a new UpdateTemplateUseCase.
func NewUpdateTemplateUseCase(templateRepo template.TemplateRepository) *UpdateTemplateUseCase {
	return &UpdateTemplateUseCase{
		templateRepo: templateRepo,
	}
}

// Execute updates a template.
func (uc *UpdateTemplateUseCase) Execute(ctx context.Context, id string, req *dtos.UpdateTemplateRequest) (*dtos.TemplateResponse, error) {
	// Validate input
	if id == "" {
		return nil, fmt.Errorf("template ID cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create template ID
	templateID, err := template.NewTemplateIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	// Find existing template
	templateEntity, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Update name if provided
	var updatedName *template.TemplateName
	if req.Name != nil {
		templateName, err := template.NewTemplateName(*req.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid template name: %w", err)
		}

		// Check if another template with same name exists
		if templateName.String() != templateEntity.Name().String() {
			exists, err := uc.templateRepo.ExistsByName(ctx, templateName)
			if err != nil {
				return nil, fmt.Errorf("failed to check template name existence: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("template with name '%s' already exists", *req.Name)
			}
		}

		updatedName = templateName
	} else {
		updatedName = templateEntity.Name()
	}

	// Update subject if provided
	var updatedSubject *template.Subject
	if req.Subject != nil {
		if *req.Subject == "" {
			updatedSubject = nil
		} else {
			subject, err := template.NewSubject(*req.Subject)
			if err != nil {
				return nil, fmt.Errorf("invalid subject: %w", err)
			}
			updatedSubject = subject
		}
	} else {
		updatedSubject = templateEntity.Subject()
	}

	// Update content if provided
	var updatedContent *template.TemplateContent
	if req.Content != nil {
		templateContent, err := template.NewTemplateContent(*req.Content)
		if err != nil {
			return nil, fmt.Errorf("invalid template content: %w", err)
		}
		updatedContent = templateContent
	} else {
		updatedContent = templateEntity.Content()
	}

	// Update tags if provided
	var updatedTags *template.Tags
	if req.Tags != nil {
		updatedTags = template.NewTags(req.Tags)
	} else {
		updatedTags = templateEntity.Tags()
	}

	// Create description (keep existing or empty)
	description := templateEntity.Description()

	// Update the template using the Update method
	if err := templateEntity.Update(
		updatedName,
		description,
		templateEntity.ChannelType(),
		updatedSubject,
		updatedContent,
		updatedTags,
	); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	// Save updated template
	if err := uc.templateRepo.Update(ctx, templateEntity); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	// Convert to response
	return dtos.ToTemplateResponse(templateEntity), nil
}