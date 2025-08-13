package template

import (
	"notification/internal/application/cqrs"
	"notification/internal/application/template/dtos"
)

// Event types
const (
	TemplateCreatedEventType = "template.created"
	TemplateUpdatedEventType = "template.updated"
	TemplateDeletedEventType = "template.deleted"
)

// TemplateCreatedEvent represents an event when a template is created
type TemplateCreatedEvent struct {
	*cqrs.BaseEvent
	Template *dtos.TemplateResponse `json:"template"`
}

// NewTemplateCreatedEvent creates a new template created event
func NewTemplateCreatedEvent(template *dtos.TemplateResponse) *TemplateCreatedEvent {
	baseEvent := cqrs.NewBaseEvent(
		TemplateCreatedEventType,
		template.ID,
		"template",
		int64(template.Version),
		template,
	)
	return &TemplateCreatedEvent{
		BaseEvent: baseEvent,
		Template:  template,
	}
}

// TemplateUpdatedEvent represents an event when a template is updated
type TemplateUpdatedEvent struct {
	*cqrs.BaseEvent
	Template *dtos.TemplateResponse `json:"template"`
}

// NewTemplateUpdatedEvent creates a new template updated event
func NewTemplateUpdatedEvent(template *dtos.TemplateResponse) *TemplateUpdatedEvent {
	baseEvent := cqrs.NewBaseEvent(
		TemplateUpdatedEventType,
		template.ID,
		"template",
		int64(template.Version),
		template,
	)
	return &TemplateUpdatedEvent{
		BaseEvent: baseEvent,
		Template:  template,
	}
}

// TemplateDeletedEvent represents an event when a template is deleted
type TemplateDeletedEvent struct {
	*cqrs.BaseEvent
	TemplateID string `json:"templateId"`
}

// NewTemplateDeletedEvent creates a new template deleted event
func NewTemplateDeletedEvent(templateID string) *TemplateDeletedEvent {
	baseEvent := cqrs.NewBaseEvent(
		TemplateDeletedEventType,
		templateID,
		"template",
		0, // Version is not critical for a deletion event
		struct{ TemplateID string }{templateID},
	)
	return &TemplateDeletedEvent{
		BaseEvent:  baseEvent,
		TemplateID: templateID,
	}
}