package template

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// TemplateID is the unique identifier for a template.
type TemplateID struct {
	value string
}

// NewTemplateID creates a new template ID.
func NewTemplateID() *TemplateID {
	return &TemplateID{
		value: "tpl_" + uuid.New().String(),
	}
}

// NewTemplateIDFromString creates a template ID from a string.
func NewTemplateIDFromString(id string) (*TemplateID, error) {
	if id == "" {
		return nil, errors.New("template ID cannot be empty")
	}
	return &TemplateID{value: id}, nil
}

// String returns the string representation.
func (t *TemplateID) String() string {
	return t.value
}

// Equals compares whether two template IDs are equal.
func (t *TemplateID) Equals(other *TemplateID) bool {
	if other == nil {
		return false
	}
	return t.value == other.value
}

// TemplateName is the name of the template.
type TemplateName struct {
	value string
}

// NewTemplateName creates a template name.
func NewTemplateName(name string) (*TemplateName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("template name cannot be empty")
	}
	if len(name) > 100 {
		return nil, errors.New("template name cannot exceed 100 characters")
	}
	return &TemplateName{value: name}, nil
}

// String returns the string representation.
func (t *TemplateName) String() string {
	return t.value
}

// Equals compares whether two template names are equal.
func (t *TemplateName) Equals(other *TemplateName) bool {
	if other == nil {
		return false
	}
	return t.value == other.value
}

// Subject is the subject of the template.
type Subject struct {
	value string
}

// NewSubject creates a subject.
func NewSubject(subject string) (*Subject, error) {
	subject = strings.TrimSpace(subject)
	if len(subject) > 200 {
		return nil, errors.New("subject cannot exceed 200 characters")
	}
	return &Subject{value: subject}, nil
}

// String returns the string representation.
func (s *Subject) String() string {
	return s.value
}

// IsEmpty checks if the subject is empty.
func (s *Subject) IsEmpty() bool {
	return s.value == ""
}

// HasVariables checks if the subject contains variables.
func (s *Subject) HasVariables() bool {
	return strings.Contains(s.value, "{") && strings.Contains(s.value, "}")
}

// ExtractVariables extracts variables from the subject.
func (s *Subject) ExtractVariables() []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(s.value, -1)

	variables := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			variable := strings.TrimSpace(match[1])
			if variable != "" && !seen[variable] {
				variables = append(variables, variable)
				seen[variable] = true
			}
		}
	}

	return variables
}

// TemplateContent is the content of the template.
type TemplateContent struct {
	value string
}

// NewTemplateContent creates template content.
func NewTemplateContent(content string) (*TemplateContent, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("template content cannot be empty")
	}
	if len(content) > 10000 {
		return nil, errors.New("template content cannot exceed 10000 characters")
	}
	return &TemplateContent{value: content}, nil
}

// String returns the string representation.
func (t *TemplateContent) String() string {
	return t.value
}

// HasVariables checks if the content contains variables.
func (t *TemplateContent) HasVariables() bool {
	return strings.Contains(t.value, "{") && strings.Contains(t.value, "}")
}

// ExtractVariables extracts variables from the content.
func (t *TemplateContent) ExtractVariables() []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(t.value, -1)

	variables := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			variable := strings.TrimSpace(match[1])
			if variable != "" && !seen[variable] {
				variables = append(variables, variable)
				seen[variable] = true
			}
		}
	}

	return variables
}

// Version is the version number.
type Version struct {
	value int
}

// NewVersion creates a version number.
func NewVersion() *Version {
	return &Version{value: 1}
}

// NewVersionFromInt creates a version number from an integer.
func NewVersionFromInt(version int) (*Version, error) {
	if version < 1 {
		return nil, errors.New("version must be positive")
	}
	return &Version{value: version}, nil
}

// Int returns the integer representation.
func (v *Version) Int() int {
	return v.value
}

// Increment increments the version number.
func (v *Version) Increment() *Version {
	return &Version{value: v.value + 1}
}

// Equals compares whether two version numbers are equal.
func (v *Version) Equals(other *Version) bool {
	if other == nil {
		return false
	}
	return v.value == other.value
}

// IsGreaterThan checks if it is greater than another version number.
func (v *Version) IsGreaterThan(other *Version) bool {
	if other == nil {
		return true
	}
	return v.value > other.value
}

// Description is the description.
type Description struct {
	value string
}

// NewDescription creates a description.
func NewDescription(desc string) (*Description, error) {
	desc = strings.TrimSpace(desc)
	if len(desc) > 500 {
		return nil, errors.New("description cannot exceed 500 characters")
	}
	return &Description{value: desc}, nil
}

// String returns the string representation.
func (d *Description) String() string {
	return d.value
}

// IsEmpty checks if the description is empty.
func (d *Description) IsEmpty() bool {
	return d.value == ""
}

// Tags are the tags.
type Tags struct {
	tags []string
}

// NewTags creates tags.
func NewTags(tags []string) *Tags {
	if tags == nil {
		tags = make([]string, 0)
	}
	// Deduplicate and sort
	uniqueTags := make([]string, 0)
	seen := make(map[string]bool)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" && !seen[tag] {
			uniqueTags = append(uniqueTags, tag)
			seen[tag] = true
		}
	}
	return &Tags{tags: uniqueTags}
}

// Add adds a tag.
func (t *Tags) Add(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}
	// Check if it already exists
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return
		}
	}
	t.tags = append(t.tags, tag)
}

// Remove removes a tag.
func (t *Tags) Remove(tag string) {
	for i, existingTag := range t.tags {
		if existingTag == tag {
			t.tags = append(t.tags[:i], t.tags[i+1:]...)
			return
		}
	}
}

// Contains checks if it contains a tag.
func (t *Tags) Contains(tag string) bool {
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// ContainsAny checks if it contains any of the tags.
func (t *Tags) ContainsAny(tags []string) bool {
	for _, tag := range tags {
		if t.Contains(tag) {
			return true
		}
	}
	return false
}

// ToSlice converts to a slice.
func (t *Tags) ToSlice() []string {
	result := make([]string, len(t.tags))
	copy(result, t.tags)
	return result
}

// Count gets the number of tags.
func (t *Tags) Count() int {
	return len(t.tags)
}
