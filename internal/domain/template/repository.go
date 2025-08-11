package template

import (
	"context"

	"channel-api/internal/domain/shared"
)

// TemplateRepository is the interface for the template repository.
type TemplateRepository interface {
	// Save saves a template.
	Save(ctx context.Context, template *Template) error
	
	// FindByID finds a template by ID.
	FindByID(ctx context.Context, id *TemplateID) (*Template, error)
	
	// FindByName finds a template by name.
	FindByName(ctx context.Context, name *TemplateName) (*Template, error)
	
	// FindAll finds all templates (supports pagination and filtering).
	FindAll(ctx context.Context, filter *TemplateFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*Template], error)
	
	// Update updates a template.
	Update(ctx context.Context, template *Template) error
	
	// Delete deletes a template.
	Delete(ctx context.Context, id *TemplateID) error
	
	// Exists checks if a template exists.
	Exists(ctx context.Context, id *TemplateID) (bool, error)
	
	// ExistsByName checks if a template with the specified name exists.
	ExistsByName(ctx context.Context, name *TemplateName) (bool, error)
}

// TemplateFilter is the filter for templates.
type TemplateFilter struct {
	ChannelType *shared.ChannelType `json:"channelType,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
}

// NewTemplateFilter creates a template filter.
func NewTemplateFilter() *TemplateFilter {
	return &TemplateFilter{}
}

// WithChannelType sets the channel type filter.
func (f *TemplateFilter) WithChannelType(channelType shared.ChannelType) *TemplateFilter {
	f.ChannelType = &channelType
	return f
}

// WithTags sets the tag filter.
func (f *TemplateFilter) WithTags(tags []string) *TemplateFilter {
	f.Tags = tags
	return f
}

// HasChannelTypeFilter checks if there is a channel type filter.
func (f *TemplateFilter) HasChannelTypeFilter() bool {
	return f.ChannelType != nil
}

// HasTagsFilter checks if there is a tag filter.
func (f *TemplateFilter) HasTagsFilter() bool {
	return len(f.Tags) > 0
}