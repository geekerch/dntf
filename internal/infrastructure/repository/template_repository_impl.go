package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"github.com/lib/pq"

	"notification/internal/domain/shared"
	"notification/internal/domain/template"
	"notification/internal/infrastructure/models"
)

// TemplateRepositoryImpl implements template.TemplateRepository interface using GORM
type TemplateRepositoryImpl struct {
	db *gorm.DB
}

// NewTemplateRepositoryImpl creates a new template repository implementation
func NewTemplateRepositoryImpl(db *gorm.DB) *TemplateRepositoryImpl {
	return &TemplateRepositoryImpl{
		db: db,
	}
}

// Save saves a template to the database
func (r *TemplateRepositoryImpl) Save(ctx context.Context, tmpl *template.Template) error {
	model, err := r.toTemplateModel(tmpl)
	if err != nil {
		return fmt.Errorf("failed to convert template to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// FindByID finds a template by its ID
func (r *TemplateRepositoryImpl) FindByID(ctx context.Context, id *template.TemplateID) (*template.Template, error) {
	var model models.TemplateModel
	
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id.String()).
		First(&model).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	return r.fromTemplateModel(&model)
}

// FindByName finds a template by its name
func (r *TemplateRepositoryImpl) FindByName(ctx context.Context, name *template.TemplateName) (*template.Template, error) {
	var model models.TemplateModel
	
	err := r.db.WithContext(ctx).
		Where("name = ? AND deleted_at IS NULL", name.String()).
		First(&model).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	return r.fromTemplateModel(&model)
}

// FindAll finds all templates with filtering and pagination
func (r *TemplateRepositoryImpl) FindAll(ctx context.Context, filter *template.TemplateFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*template.Template], error) {
	query := r.db.WithContext(ctx).Model(&models.TemplateModel{}).Where("deleted_at IS NULL")

	// Apply filters
	if filter.HasChannelTypeFilter() {
		query = query.Where("channel_type = ?", filter.ChannelType.String())
	}

	if filter.HasTagsFilter() {
		// For PostgreSQL, use array overlap operator
		if r.db.Dialector.Name() == "postgres" {
			query = query.Where("tags && ?", pq.StringArray(filter.Tags))
		} else {
			// For other databases, use JSON contains logic
			for _, tag := range filter.Tags {
				query = query.Where("JSON_EXTRACT(tags, '$') LIKE ?", "%"+tag+"%")
			}
		}
	}

	// Count total records
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count templates: %w", err)
	}

	// Query templates with pagination
	var templateModels []models.TemplateModel
	err := query.
		Order("created_at DESC").
		Limit(pagination.MaxResultCount).
		Offset(pagination.SkipCount).
		Find(&templateModels).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}

	// Convert to domain objects
	templates := make([]*template.Template, 0, len(templateModels))
	for _, model := range templateModels {
		tmpl, err := r.fromTemplateModel(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to template: %w", err)
		}
		templates = append(templates, tmpl)
	}

	// Calculate hasMore
	hasMore := pagination.SkipCount+len(templates) < int(totalCount)

	return &shared.PaginatedResult[*template.Template]{
		Items:          templates,
		SkipCount:      pagination.SkipCount,
		MaxResultCount: pagination.MaxResultCount,
		TotalCount:     int(totalCount),
		HasMore:        hasMore,
	}, nil
}

// Update updates a template in the database
func (r *TemplateRepositoryImpl) Update(ctx context.Context, tmpl *template.Template) error {
	model, err := r.toTemplateModel(tmpl)
	if err != nil {
		return fmt.Errorf("failed to convert template to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

// Delete deletes a template from the database (hard delete)
func (r *TemplateRepositoryImpl) Delete(ctx context.Context, id *template.TemplateID) error {
	if err := r.db.WithContext(ctx).Delete(&models.TemplateModel{}, "id = ?", id.String()).Error; err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// Exists checks if a template exists
func (r *TemplateRepositoryImpl) Exists(ctx context.Context, id *template.TemplateID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TemplateModel{}).
		Where("id = ? AND deleted_at IS NULL", id.String()).
		Count(&count).Error
	
	if err != nil {
		return false, fmt.Errorf("failed to check template existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if a template with the given name exists
func (r *TemplateRepositoryImpl) ExistsByName(ctx context.Context, name *template.TemplateName) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TemplateModel{}).
		Where("name = ? AND deleted_at IS NULL", name.String()).
		Count(&count).Error
	
	if err != nil {
		return false, fmt.Errorf("failed to check template name existence: %w", err)
	}

	return count > 0, nil
}

// toTemplateModel converts domain template to GORM model
func (r *TemplateRepositoryImpl) toTemplateModel(tmpl *template.Template) (*models.TemplateModel, error) {
	// Handle deleted_at
	var deletedAt *int64
	if tmpl.Timestamps().DeletedAt != nil {
		deletedAt = tmpl.Timestamps().DeletedAt
	}

	return &models.TemplateModel{
		ID:          tmpl.ID().String(),
		Name:        tmpl.Name().String(),
		Description: tmpl.Description().String(),
		ChannelType: tmpl.ChannelType().String(),
		Subject:     tmpl.Subject().String(),
		Content:     tmpl.Content().String(),
		Tags:        pq.StringArray(tmpl.Tags().ToSlice()),
		CreatedAt:   tmpl.Timestamps().CreatedAt,
		UpdatedAt:   tmpl.Timestamps().UpdatedAt,
		DeletedAt:   deletedAt,
		Version:     tmpl.Version().Int(),
	}, nil
}

// fromTemplateModel converts GORM model to domain template
func (r *TemplateRepositoryImpl) fromTemplateModel(model *models.TemplateModel) (*template.Template, error) {
	// Convert ID
	id, err := template.NewTemplateIDFromString(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	// Convert name
	name, err := template.NewTemplateName(model.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	// Convert description
	description, err := template.NewDescription(model.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	// Convert channel type
	channelType, err := shared.NewChannelTypeFromString(model.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("invalid channel type: %s, error: %w", model.ChannelType, err)
	}
	if !channelType.IsValid() {
		return nil, fmt.Errorf("invalid channel type: %s", model.ChannelType)
	}

	// Convert subject
	subject, err := template.NewSubject(model.Subject)
	if err != nil {
		return nil, fmt.Errorf("invalid subject: %w", err)
	}

	// Convert content
	content, err := template.NewTemplateContent(model.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid content: %w", err)
	}

	// Convert tags
	tags := template.NewTags(model.Tags)

	// Convert version
	version, err := template.NewVersionFromInt(model.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	// Convert timestamps
	timestamps := &shared.Timestamps{
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		DeletedAt: model.DeletedAt,
	}

	// Reconstruct template
	return template.ReconstructTemplate(
		id,
		name,
		description,
		channelType,
		subject,
		content,
		tags,
		timestamps,
		version,
	), nil
}