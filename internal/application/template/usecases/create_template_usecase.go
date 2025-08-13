package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/template/dtos"
	"notification/internal/domain/template"
)

// CreateTemplateUseCase handles the creation of templates.
type CreateTemplateUseCase struct {
	templateRepo template.TemplateRepository
}

// NewCreateTemplateUseCase creates a new CreateTemplateUseCase.
func NewCreateTemplateUseCase(templateRepo template.TemplateRepository) *CreateTemplateUseCase {
	return &CreateTemplateUseCase{
		templateRepo: templateRepo,
	}
}

// Execute creates a new template.
func (uc *CreateTemplateUseCase) Execute(ctx context.Context, req *dtos.CreateTemplateRequest) (*dtos.TemplateResponse, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create template name
	templateName, err := template.NewTemplateName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	// Check if template with same name already exists
	exists, err := uc.templateRepo.ExistsByName(ctx, templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to check template existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("template with name '%s' already exists", req.Name)
	}

	// Create template content
	templateContent, err := template.NewTemplateContent(req.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid template content: %w", err)
	}

	// Create subject if provided
	var subject *template.Subject
	if req.Subject != "" {
		subject, err = template.NewSubject(req.Subject)
		if err != nil {
			return nil, fmt.Errorf("invalid subject: %w", err)
		}
	}

	// Create description (empty for now)
	description, err := template.NewDescription("")
	if err != nil {
		return nil, fmt.Errorf("failed to create description: %w", err)
	}

	// Create tags
	tags := template.NewTags(req.Tags)

	// Create template entity
	templateEntity, err := template.NewTemplate(
		templateName,
		description,
		req.ChannelType,
		subject,
		templateContent,
		tags,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	// Save template
	if err := uc.templateRepo.Save(ctx, templateEntity); err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	// Convert to response
	return dtos.ToTemplateResponse(templateEntity), nil
}