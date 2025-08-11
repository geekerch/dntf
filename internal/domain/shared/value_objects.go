package shared

import (
	"errors"
	"time"
)

// ChannelType represents the type of communication channel
type ChannelType string

const (
	ChannelTypeEmail ChannelType = "email"
	ChannelTypeSlack ChannelType = "slack"
	ChannelTypeSMS   ChannelType = "sms"
)

// IsValid validates if the channel type is valid
func (ct ChannelType) IsValid() bool {
	switch ct {
	case ChannelTypeEmail, ChannelTypeSlack, ChannelTypeSMS:
		return true
	default:
		return false
	}
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