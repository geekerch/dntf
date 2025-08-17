package template

import (
	"errors"

	"notification/internal/domain/shared"
)

// Template is the aggregate root for templates.
type Template struct {
	id          *TemplateID
	name        *TemplateName
	description *Description
	channelType shared.ChannelType
	subject     *Subject
	content     *TemplateContent
	tags        *Tags
	timestamps  *shared.Timestamps
	version     *Version
}

// NewTemplate creates a new template.
func NewTemplate(
	name *TemplateName,
	description *Description,
	channelType shared.ChannelType,
	subject *Subject,
	content *TemplateContent,
	tags *Tags,
) (*Template, error) {
	// Validate required fields
	if name == nil {
		return nil, errors.New("template name is required")
	}
	if !channelType.IsValid() {
		return nil, errors.New("invalid channel type")
	}
	if content == nil {
		return nil, errors.New("template content is required")
	}

	// Set default values
	if description == nil {
		description, _ = NewDescription("")
	}
	if subject == nil {
		subject, _ = NewSubject("")
	}
	if tags == nil {
		tags = NewTags(nil)
	}

	return &Template{
		id:          NewTemplateID(),
		name:        name,
		description: description,
		channelType: channelType,
		subject:     subject,
		content:     content,
		tags:        tags,
		timestamps:  shared.NewTimestamps(),
		version:     NewVersion(),
	}, nil
}

// ReconstructTemplate reconstructs a template from persistent data.
func ReconstructTemplate(
	id *TemplateID,
	name *TemplateName,
	description *Description,
	channelType shared.ChannelType,
	subject *Subject,
	content *TemplateContent,
	tags *Tags,
	timestamps *shared.Timestamps,
	version *Version,
) *Template {
	return &Template{
		id:          id,
		name:        name,
		description: description,
		channelType: channelType,
		subject:     subject,
		content:     content,
		tags:        tags,
		timestamps:  timestamps,
		version:     version,
	}
}

// ID gets the template ID.
func (t *Template) ID() *TemplateID {
	return t.id
}

// Name gets the template name.
func (t *Template) Name() *TemplateName {
	return t.name
}

// Description gets the description.
func (t *Template) Description() *Description {
	return t.description
}

// ChannelType gets the channel type.
func (t *Template) ChannelType() shared.ChannelType {
	return t.channelType
}

// Subject gets the subject.
func (t *Template) Subject() *Subject {
	return t.subject
}

// Content gets the template content.
func (t *Template) Content() *TemplateContent {
	return t.content
}

// Tags gets the tags.
func (t *Template) Tags() *Tags {
	return t.tags
}

// Timestamps gets the timestamps.
func (t *Template) Timestamps() *shared.Timestamps {
	return t.timestamps
}

// Version gets the version number.
func (t *Template) Version() *Version {
	return t.version
}

// Update updates the template.
func (t *Template) Update(
	name *TemplateName,
	description *Description,
	channelType shared.ChannelType,
	subject *Subject,
	content *TemplateContent,
	tags *Tags,
) error {
	// Validate required fields
	if name == nil {
		return errors.New("template name is required")
	}
	if !channelType.IsValid() {
		return errors.New("invalid channel type")
	}
	if content == nil {
		return errors.New("template content is required")
	}

	// Set default values
	if description == nil {
		description, _ = NewDescription("")
	}
	if subject == nil {
		subject, _ = NewSubject("")
	}
	if tags == nil {
		tags = NewTags(nil)
	}

	// Update fields
	t.name = name
	t.description = description
	t.channelType = channelType
	t.subject = subject
	t.content = content
	t.tags = tags
	t.timestamps.UpdateTimestamp()
	t.version = t.version.Increment()

	return nil
}

// Delete soft deletes the template.
func (t *Template) Delete() error {
	if t.timestamps.IsDeleted() {
		return errors.New("template is already deleted")
	}
	t.timestamps.MarkDeleted()
	return nil
}

// IsDeleted checks if the template is deleted.
func (t *Template) IsDeleted() bool {
	return t.timestamps.IsDeleted()
}

// HasTag checks if it contains the specified tag.
func (t *Template) HasTag(tag string) bool {
	return t.tags.Contains(tag)
}

// HasAnyTag checks if it contains any of the specified tags.
func (t *Template) HasAnyTag(tags []string) bool {
	return t.tags.ContainsAny(tags)
}

// MatchesType checks if the channel type matches.
func (t *Template) MatchesType(channelType shared.ChannelType) bool {
	return t.channelType == channelType
}

// GetAllVariables gets all variables in the template.
func (t *Template) GetAllVariables() []string {
	variables := make(map[string]bool)

	// Extract variables from the subject
	for _, variable := range t.subject.ExtractVariables() {
		variables[variable] = true
	}

	// Extract variables from the content
	for _, variable := range t.content.ExtractVariables() {
		variables[variable] = true
	}

	// Convert to slice
	result := make([]string, 0, len(variables))
	for variable := range variables {
		result = append(result, variable)
	}

	return result
}

// ValidateVariables validates if the provided variables contain all the required variables for the template.
func (t *Template) ValidateVariables(providedVariables map[string]interface{}) []string {
	requiredVariables := t.GetAllVariables()
	missingVariables := make([]string, 0)

	for _, variable := range requiredVariables {
		if _, exists := providedVariables[variable]; !exists {
			missingVariables = append(missingVariables, variable)
		}
	}

	return missingVariables
}
