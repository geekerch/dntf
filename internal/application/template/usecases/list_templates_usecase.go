package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/template/dtos"
	"notification/internal/domain/template"
)

// ListTemplatesUseCase handles listing templates.
type ListTemplatesUseCase struct {
	templateRepo template.TemplateRepository
}

// NewListTemplatesUseCase creates a new ListTemplatesUseCase.
func NewListTemplatesUseCase(templateRepo template.TemplateRepository) *ListTemplatesUseCase {
	return &ListTemplatesUseCase{
		templateRepo: templateRepo,
	}
}

// Execute lists templates with filtering and pagination.
func (uc *ListTemplatesUseCase) Execute(ctx context.Context, req *dtos.ListTemplatesRequest) (*dtos.ListTemplatesResponse, error) {
	// Use default request if nil
	if req == nil {
		req = &dtos.ListTemplatesRequest{}
	}

	// Convert to filter and pagination
	filter := req.ToTemplateFilter()
	pagination := req.ToPagination()

	// Find templates
	result, err := uc.templateRepo.FindAll(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to find templates: %w", err)
	}

	// Convert to response
	return &dtos.ListTemplatesResponse{
		Items:          dtos.ToTemplateResponseList(result.Items),
		SkipCount:      result.SkipCount,
		MaxResultCount: result.MaxResultCount,
		TotalCount:     result.TotalCount,
		HasMore:        result.HasMore,
	}, nil
}