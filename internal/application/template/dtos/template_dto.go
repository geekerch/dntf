package dtos

import (
	"time"

	"notification/internal/domain/shared"
	"notification/internal/domain/template"
)

// CreateTemplateRequest represents the request to create a template.
type CreateTemplateRequest struct {
	Name        string                `json:"name" validate:"required,min=1,max=100"`
	ChannelType shared.ChannelType    `json:"channelType" validate:"required"`
	Subject     string                `json:"subject,omitempty" validate:"max=200"`
	Content     string                `json:"content" validate:"required"`
	Variables   []string              `json:"variables,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Settings    *shared.CommonSettings `json:"settings,omitempty"`
}

// UpdateTemplateRequest represents the request to update a template.
type UpdateTemplateRequest struct {
	Name        *string               `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Subject     *string               `json:"subject,omitempty" validate:"omitempty,max=200"`
	Content     *string               `json:"content,omitempty" validate:"omitempty,min=1"`
	Variables   []string              `json:"variables,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Settings    *shared.CommonSettings `json:"settings,omitempty"`
}

// TemplateResponse represents the response for a template.
type TemplateResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	ChannelType shared.ChannelType    `json:"channelType"`
	Subject     string                `json:"subject,omitempty"`
	Content     string                `json:"content"`
	Variables   []string              `json:"variables,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Version     int                   `json:"version"`
	Settings    *shared.CommonSettings `json:"settings,omitempty"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

// ListTemplatesRequest represents the request to list templates.
type ListTemplatesRequest struct {
	ChannelType    *shared.ChannelType `json:"channelType,omitempty"`
	Tags           []string            `json:"tags,omitempty"`
	SkipCount      int                 `json:"skipCount,omitempty" validate:"omitempty,min=0"`
	MaxResultCount int                 `json:"maxResultCount,omitempty" validate:"omitempty,min=1,max=100"`
}

// ListTemplatesResponse represents the response for listing templates.
type ListTemplatesResponse struct {
	Items          []*TemplateResponse `json:"items"`
	SkipCount      int                 `json:"skipCount"`
	MaxResultCount int                 `json:"maxResultCount"`
	TotalCount     int                 `json:"totalCount"`
	HasMore        bool                `json:"hasMore"`
}

// ToTemplateResponse converts a template entity to a response DTO.
func ToTemplateResponse(t *template.Template) *TemplateResponse {
	if t == nil {
		return nil
	}

	// Convert timestamps from Unix milliseconds to time.Time
	createdAt := time.Unix(0, t.Timestamps().CreatedAt*int64(time.Millisecond))
	updatedAt := time.Unix(0, t.Timestamps().UpdatedAt*int64(time.Millisecond))

	response := &TemplateResponse{
		ID:          t.ID().String(),
		Name:        t.Name().String(),
		ChannelType: t.ChannelType(),
		Content:     t.Content().String(),
		Variables:   t.GetAllVariables(),
		Tags:        t.Tags().ToSlice(),
		Version:     t.Version().Int(),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if t.Subject() != nil && !t.Subject().IsEmpty() {
		response.Subject = t.Subject().String()
	}

	return response
}

// ToTemplateResponseList converts a list of template entities to response DTOs.
func ToTemplateResponseList(templates []*template.Template) []*TemplateResponse {
	responses := make([]*TemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = ToTemplateResponse(t)
	}
	return responses
}

// ToTemplateFilter converts a list request to a template filter.
func (req *ListTemplatesRequest) ToTemplateFilter() *template.TemplateFilter {
	filter := template.NewTemplateFilter()
	
	if req.ChannelType != nil {
		filter.WithChannelType(*req.ChannelType)
	}
	
	if len(req.Tags) > 0 {
		filter.WithTags(req.Tags)
	}
	
	return filter
}

// ToPagination converts a list request to pagination.
func (req *ListTemplatesRequest) ToPagination() *shared.Pagination {
	skipCount := req.SkipCount
	maxResultCount := req.MaxResultCount
	
	// Set defaults if not provided
	if maxResultCount <= 0 {
		maxResultCount = 20
	}
	if skipCount < 0 {
		skipCount = 0
	}
	
	pagination, err := shared.NewPagination(skipCount, maxResultCount)
	if err != nil {
		// Return default pagination if there's an error
		return shared.DefaultPagination()
	}
	
	return pagination
}