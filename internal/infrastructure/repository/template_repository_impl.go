package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"channel-api/internal/domain/shared"
	"channel-api/internal/domain/template"
)

// TemplateRepositoryImpl implements template.TemplateRepository interface
type TemplateRepositoryImpl struct {
	db *sqlx.DB
}

// NewTemplateRepositoryImpl creates a new template repository implementation
func NewTemplateRepositoryImpl(db *sqlx.DB) *TemplateRepositoryImpl {
	return &TemplateRepositoryImpl{
		db: db,
	}
}

// templateRow represents template data in database
type templateRow struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	ChannelType string         `db:"channel_type"`
	Subject     string         `db:"subject"`
	Content     string         `db:"content"`
	Tags        pq.StringArray `db:"tags"`
	CreatedAt   int64          `db:"created_at"`
	UpdatedAt   int64          `db:"updated_at"`
	DeletedAt   sql.NullInt64  `db:"deleted_at"`
	Version     int            `db:"version"`
}

// Save saves a template to the database
func (r *TemplateRepositoryImpl) Save(ctx context.Context, tmpl *template.Template) error {
	row, err := r.toTemplateRow(tmpl)
	if err != nil {
		return fmt.Errorf("failed to convert template to row: %w", err)
	}

	query := `
		INSERT INTO templates (
			id, name, description, channel_type, subject, content, tags,
			created_at, updated_at, deleted_at, version
		) VALUES (
			:id, :name, :description, :channel_type, :subject, :content, :tags,
			:created_at, :updated_at, :deleted_at, :version
		)`

	_, err = r.db.NamedExecContext(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// FindByID finds a template by its ID
func (r *TemplateRepositoryImpl) FindByID(ctx context.Context, id *template.TemplateID) (*template.Template, error) {
	query := `
		SELECT id, name, description, channel_type, subject, content, tags,
			   created_at, updated_at, deleted_at, version
		FROM templates 
		WHERE id = $1 AND deleted_at IS NULL`

	var row templateRow
	err := r.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	return r.fromTemplateRow(&row)
}

// FindByName finds a template by its name
func (r *TemplateRepositoryImpl) FindByName(ctx context.Context, name *template.TemplateName) (*template.Template, error) {
	query := `
		SELECT id, name, description, channel_type, subject, content, tags,
			   created_at, updated_at, deleted_at, version
		FROM templates 
		WHERE name = $1 AND deleted_at IS NULL`

	var row templateRow
	err := r.db.GetContext(ctx, &row, query, name.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	return r.fromTemplateRow(&row)
}

// FindAll finds all templates with filtering and pagination
func (r *TemplateRepositoryImpl) FindAll(ctx context.Context, filter *template.TemplateFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*template.Template], error) {
	// Build WHERE clause
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to build where clause: %w", err)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM templates %s", whereClause)
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to count templates: %w", err)
	}

	// Query templates with pagination
	query := fmt.Sprintf(`
		SELECT id, name, description, channel_type, subject, content, tags,
			   created_at, updated_at, deleted_at, version
		FROM templates %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)

	args = append(args, pagination.MaxResultCount, pagination.SkipCount)

	var rows []templateRow
	err = r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}

	// Convert to domain objects
	templates := make([]*template.Template, 0, len(rows))
	for _, row := range rows {
		tmpl, err := r.fromTemplateRow(&row)
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to template: %w", err)
		}
		templates = append(templates, tmpl)
	}

	// Calculate hasMore
	hasMore := pagination.SkipCount+len(templates) < totalCount

	return &shared.PaginatedResult[*template.Template]{
		Items:          templates,
		SkipCount:      pagination.SkipCount,
		MaxResultCount: pagination.MaxResultCount,
		TotalCount:     totalCount,
		HasMore:        hasMore,
	}, nil
}

// Update updates a template in the database
func (r *TemplateRepositoryImpl) Update(ctx context.Context, tmpl *template.Template) error {
	row, err := r.toTemplateRow(tmpl)
	if err != nil {
		return fmt.Errorf("failed to convert template to row: %w", err)
	}

	query := `
		UPDATE templates SET
			name = :name,
			description = :description,
			channel_type = :channel_type,
			subject = :subject,
			content = :content,
			tags = :tags,
			updated_at = :updated_at,
			deleted_at = :deleted_at,
			version = :version
		WHERE id = :id`

	_, err = r.db.NamedExecContext(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

// Delete deletes a template from the database (hard delete)
func (r *TemplateRepositoryImpl) Delete(ctx context.Context, id *template.TemplateID) error {
	query := `DELETE FROM templates WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// Exists checks if a template exists
func (r *TemplateRepositoryImpl) Exists(ctx context.Context, id *template.TemplateID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM templates WHERE id = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id.String())
	if err != nil {
		return false, fmt.Errorf("failed to check template existence: %w", err)
	}

	return exists, nil
}

// ExistsByName checks if a template with the given name exists
func (r *TemplateRepositoryImpl) ExistsByName(ctx context.Context, name *template.TemplateName) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM templates WHERE name = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, name.String())
	if err != nil {
		return false, fmt.Errorf("failed to check template name existence: %w", err)
	}

	return exists, nil
}

// buildWhereClause builds WHERE clause for filtering
func (r *TemplateRepositoryImpl) buildWhereClause(filter *template.TemplateFilter) (string, []interface{}, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if filter.HasChannelTypeFilter() {
		conditions = append(conditions, fmt.Sprintf("channel_type = $%d", argIndex))
		args = append(args, string(*filter.ChannelType))
		argIndex++
	}

	if filter.HasTagsFilter() {
		conditions = append(conditions, fmt.Sprintf("tags && $%d", argIndex))
		args = append(args, pq.Array(filter.Tags))
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args, nil
}

// toTemplateRow converts domain template to database row
func (r *TemplateRepositoryImpl) toTemplateRow(tmpl *template.Template) (*templateRow, error) {
	// Handle deleted_at
	var deletedAt sql.NullInt64
	if tmpl.Timestamps().DeletedAt != nil {
		deletedAt = sql.NullInt64{
			Int64: *tmpl.Timestamps().DeletedAt,
			Valid: true,
		}
	}

	return &templateRow{
		ID:          tmpl.ID().String(),
		Name:        tmpl.Name().String(),
		Description: tmpl.Description().String(),
		ChannelType: string(tmpl.ChannelType()),
		Subject:     tmpl.Subject().String(),
		Content:     tmpl.Content().String(),
		Tags:        pq.StringArray(tmpl.Tags().ToSlice()),
		CreatedAt:   tmpl.Timestamps().CreatedAt,
		UpdatedAt:   tmpl.Timestamps().UpdatedAt,
		DeletedAt:   deletedAt,
		Version:     tmpl.Version().Int(),
	}, nil
}

// fromTemplateRow converts database row to domain template
func (r *TemplateRepositoryImpl) fromTemplateRow(row *templateRow) (*template.Template, error) {
	// Convert ID
	id, err := template.NewTemplateIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid template ID: %w", err)
	}

	// Convert name
	name, err := template.NewTemplateName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	// Convert description
	description, err := template.NewDescription(row.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	// Convert channel type
	channelType := shared.ChannelType(row.ChannelType)
	if !channelType.IsValid() {
		return nil, fmt.Errorf("invalid channel type: %s", row.ChannelType)
	}

	// Convert subject
	subject, err := template.NewSubject(row.Subject)
	if err != nil {
		return nil, fmt.Errorf("invalid subject: %w", err)
	}

	// Convert content
	content, err := template.NewTemplateContent(row.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid content: %w", err)
	}

	// Convert tags
	tags := template.NewTags([]string(row.Tags))

	// Convert version
	version, err := template.NewVersionFromInt(row.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	// Convert timestamps
	timestamps := &shared.Timestamps{
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		timestamps.DeletedAt = &row.DeletedAt.Int64
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