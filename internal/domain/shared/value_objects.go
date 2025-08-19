package shared

import (
	"errors"
	"fmt"
	"time"
)

// ChannelType represents the type of communication channel
type ChannelType struct {
	name string
}

// Predefined channel types for backward compatibility
var (
	ChannelTypeEmail = MustNewChannelType("email")
	ChannelTypeSlack = MustNewChannelType("slack")
	ChannelTypeSMS   = MustNewChannelType("sms")
)

// NewChannelType creates a new channel type
func NewChannelType(name string) (ChannelType, error) {
	if name == "" {
		return ChannelType{}, errors.New("channel type name cannot be empty")
	}
	
	// Check if the channel type is registered
	if !GetChannelTypeRegistry().IsValidChannelType(name) {
		return ChannelType{}, fmt.Errorf("invalid channel type: %s", name)
	}
	
	return ChannelType{name: name}, nil
}

// MustNewChannelType creates a new channel type and panics if invalid
func MustNewChannelType(name string) ChannelType {
	ct, err := NewChannelType(name)
	if err != nil {
		// For predefined types, we allow creation even if not registered yet
		// This is to avoid circular dependency during initialization
		return ChannelType{name: name}
	}
	return ct
}

// NewChannelTypeFromString creates a channel type from string (for backward compatibility)
func NewChannelTypeFromString(name string) (ChannelType, error) {
	return NewChannelType(name)
}

// String returns the string representation of the channel type
func (ct ChannelType) String() string {
	return ct.name
}

// IsValid validates if the channel type is valid
func (ct ChannelType) IsValid() bool {
	if ct.name == "" {
		return false
	}
	return GetChannelTypeRegistry().IsValidChannelType(ct.name)
}

// Equals compares two channel types for equality
func (ct ChannelType) Equals(other ChannelType) bool {
	return ct.name == other.name
}

// Pagination represents pagination parameters
type Pagination struct {
	SkipCount      int `json:"skipCount"`
	MaxResultCount int `json:"maxResultCount"`
}

// NewPagination creates new pagination parameters
func NewPagination(skipCount, maxResultCount int) (*Pagination, error) {
	if skipCount < 0 {
		return nil, errors.New("skipCount must be non-negative")
	}
	if maxResultCount < 1 || maxResultCount > 100 {
		return nil, errors.New("maxResultCount must be between 1 and 100")
	}
	return &Pagination{
		SkipCount:      skipCount,
		MaxResultCount: maxResultCount,
	}, nil
}

// DefaultPagination returns default pagination parameters
func DefaultPagination() *Pagination {
	return &Pagination{
		SkipCount:      0,
		MaxResultCount: 10,
	}
}

// PaginatedResult represents paginated query result
type PaginatedResult[T any] struct {
	Items          []T  `json:"items"`
	SkipCount      int  `json:"skipCount"`
	MaxResultCount int  `json:"maxResultCount"`
	TotalCount     int  `json:"totalCount"`
	HasMore        bool `json:"hasMore"`
}

// CommonSettings represents common configuration settings
type CommonSettings struct {
	Timeout       int `json:"timeout"`       // timeout in milliseconds
	RetryAttempts int `json:"retryAttempts"` // number of retry attempts
	RetryDelay    int `json:"retryDelay"`    // retry delay in milliseconds
}

// NewCommonSettings creates new common settings
func NewCommonSettings(timeout, retryAttempts, retryDelay int) (*CommonSettings, error) {
	if timeout <= 0 {
		return nil, errors.New("timeout must be positive")
	}
	if retryAttempts < 0 {
		return nil, errors.New("retryAttempts must be non-negative")
	}
	if retryDelay < 0 {
		return nil, errors.New("retryDelay must be non-negative")
	}
	
	return &CommonSettings{
		Timeout:       timeout,
		RetryAttempts: retryAttempts,
		RetryDelay:    retryDelay,
	}, nil
}

// Timestamps represents creation, update, and deletion timestamps
type Timestamps struct {
	CreatedAt int64  `json:"createdAt"` // Unix timestamp in milliseconds
	UpdatedAt int64  `json:"updatedAt"` // Unix timestamp in milliseconds
	DeletedAt *int64 `json:"deletedAt,omitempty"` // Unix timestamp in milliseconds
}

// NewTimestamps creates new timestamps
func NewTimestamps() *Timestamps {
	now := time.Now().UnixMilli()
	return &Timestamps{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateTimestamp updates the timestamp to current time
func (t *Timestamps) UpdateTimestamp() {
	t.UpdatedAt = time.Now().UnixMilli()
}

// MarkDeleted marks the entity as deleted
func (t *Timestamps) MarkDeleted() {
	now := time.Now().UnixMilli()
	t.DeletedAt = &now
}

// IsDeleted checks if the entity is deleted
func (t *Timestamps) IsDeleted() bool {
	return t.DeletedAt != nil
}

// MarshalJSON implements json.Marshaler interface for ChannelType
func (ct ChannelType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ct.name + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler interface for ChannelType
func (ct *ChannelType) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid JSON string for ChannelType: %s", string(data))
	}
	
	name := string(data[1 : len(data)-1])
	if name == "" {
		return errors.New("channel type name cannot be empty")
	}
	
	// Create channel type (this will validate against registry)
	channelType, err := NewChannelType(name)
	if err != nil {
		return err
	}
	
	*ct = channelType
	return nil
}