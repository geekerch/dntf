package channel

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// ChannelID represents a unique channel identifier
type ChannelID struct {
	value string
}

// NewChannelID creates a new channel ID
func NewChannelID() *ChannelID {
	return &ChannelID{
		value: "channel_" + uuid.New().String(),
	}
}

// NewChannelIDFromString creates a channel ID from string
func NewChannelIDFromString(id string) (*ChannelID, error) {
	if id == "" {
		return nil, errors.New("channel ID cannot be empty")
	}
	return &ChannelID{value: id}, nil
}

// String returns string representation
func (c *ChannelID) String() string {
	return c.value
}

// Equals compares two channel IDs for equality
func (c *ChannelID) Equals(other *ChannelID) bool {
	if other == nil {
		return false
	}
	return c.value == other.value
}

// ChannelName represents a channel name
type ChannelName struct {
	value string
}

// NewChannelName creates a new channel name
func NewChannelName(name string) (*ChannelName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("channel name cannot be empty")
	}
	if len(name) > 100 {
		return nil, errors.New("channel name cannot exceed 100 characters")
	}
	// Check name format (letters, numbers, underscores, hyphens)
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return nil, errors.New("channel name can only contain letters, numbers, underscores, and hyphens")
	}
	return &ChannelName{value: name}, nil
}

// String returns string representation
func (c *ChannelName) String() string {
	return c.value
}

// Equals compares two channel names for equality
func (c *ChannelName) Equals(other *ChannelName) bool {
	if other == nil {
		return false
	}
	return c.value == other.value
}

// Description represents a description
type Description struct {
	value string
}

// NewDescription creates a new description
func NewDescription(desc string) (*Description, error) {
	desc = strings.TrimSpace(desc)
	if len(desc) > 500 {
		return nil, errors.New("description cannot exceed 500 characters")
	}
	return &Description{value: desc}, nil
}

// String returns string representation
func (d *Description) String() string {
	return d.value
}

// IsEmpty checks if description is empty
func (d *Description) IsEmpty() bool {
	return d.value == ""
}

// ChannelConfig represents channel configuration
type ChannelConfig struct {
	data map[string]interface{}
}

// NewChannelConfig creates a new channel configuration
func NewChannelConfig(config map[string]interface{}) *ChannelConfig {
	if config == nil {
		config = make(map[string]interface{})
	}
	return &ChannelConfig{data: config}
}

// Get retrieves configuration value
func (c *ChannelConfig) Get(key string) (interface{}, bool) {
	value, exists := c.data[key]
	return value, exists
}

// Set sets configuration value
func (c *ChannelConfig) Set(key string, value interface{}) {
	c.data[key] = value
}

// ToMap converts to map
func (c *ChannelConfig) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Recipient represents a message recipient
type Recipient struct {
	Name   string `json:"name"`
	Email  string `json:"email,omitempty"`
	Target string `json:"target,omitempty"`
	Type   string `json:"type"`
}

// NewRecipient creates a new recipient
func NewRecipient(name, email, target, recipientType string) (*Recipient, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("recipient name cannot be empty")
	}
	
	recipientType = strings.TrimSpace(recipientType)
	if recipientType == "" {
		return nil, errors.New("recipient type cannot be empty")
	}
	
	return &Recipient{
		Name:   name,
		Email:  email,
		Target: target,
		Type:   recipientType,
	}, nil
}

// Recipients represents a list of recipients
type Recipients struct {
	recipients []*Recipient
}

// NewRecipients creates a new recipients list
func NewRecipients(recipients []*Recipient) *Recipients {
	if recipients == nil {
		recipients = make([]*Recipient, 0)
	}
	return &Recipients{recipients: recipients}
}

// Add adds a recipient to the list
func (r *Recipients) Add(recipient *Recipient) error {
	if recipient == nil {
		return errors.New("recipient cannot be nil")
	}
	r.recipients = append(r.recipients, recipient)
	return nil
}

// ToSlice converts to slice
func (r *Recipients) ToSlice() []*Recipient {
	result := make([]*Recipient, len(r.recipients))
	copy(result, r.recipients)
	return result
}

// Count returns the number of recipients
func (r *Recipients) Count() int {
	return len(r.recipients)
}

// Tags represents a collection of tags
type Tags struct {
	tags []string
}

// NewTags creates a new tags collection
func NewTags(tags []string) *Tags {
	if tags == nil {
		tags = make([]string, 0)
	}
	// Remove duplicates and empty strings
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

// Add adds a tag to the collection
func (t *Tags) Add(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}
	// Check if tag already exists
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return
		}
	}
	t.tags = append(t.tags, tag)
}

// Remove removes a tag from the collection
func (t *Tags) Remove(tag string) {
	for i, existingTag := range t.tags {
		if existingTag == tag {
			t.tags = append(t.tags[:i], t.tags[i+1:]...)
			return
		}
	}
}

// Contains checks if the collection contains the specified tag
func (t *Tags) Contains(tag string) bool {
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// ContainsAny checks if the collection contains any of the specified tags
func (t *Tags) ContainsAny(tags []string) bool {
	for _, tag := range tags {
		if t.Contains(tag) {
			return true
		}
	}
	return false
}

// ToSlice converts to slice
func (t *Tags) ToSlice() []string {
	result := make([]string, len(t.tags))
	copy(result, t.tags)
	return result
}

// Count returns the number of tags
func (t *Tags) Count() int {
	return len(t.tags)
}