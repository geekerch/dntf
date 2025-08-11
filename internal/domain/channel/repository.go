package channel

import (
	"context"

	"channel-api/internal/domain/shared"
)

// ChannelRepository 通道倉儲介面
type ChannelRepository interface {
	// Save 儲存通道
	Save(ctx context.Context, channel *Channel) error
	
	// FindByID 根據 ID 查找通道
	FindByID(ctx context.Context, id *ChannelID) (*Channel, error)
	
	// FindByName 根據名稱查找通道
	FindByName(ctx context.Context, name *ChannelName) (*Channel, error)
	
	// FindAll 查找所有通道 (支援分頁和過濾)
	FindAll(ctx context.Context, filter *ChannelFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*Channel], error)
	
	// Update 更新通道
	Update(ctx context.Context, channel *Channel) error
	
	// Delete 刪除通道
	Delete(ctx context.Context, id *ChannelID) error
	
	// Exists 檢查通道是否存在
	Exists(ctx context.Context, id *ChannelID) (bool, error)
	
	// ExistsByName 檢查指定名稱的通道是否存在
	ExistsByName(ctx context.Context, name *ChannelName) (bool, error)
}

// ChannelFilter 通道過濾條件
type ChannelFilter struct {
	ChannelType *shared.ChannelType `json:"channelType,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Enabled     *bool               `json:"enabled,omitempty"`
}

// NewChannelFilter 建立通道過濾條件
func NewChannelFilter() *ChannelFilter {
	return &ChannelFilter{}
}

// WithChannelType 設定通道類型過濾
func (f *ChannelFilter) WithChannelType(channelType shared.ChannelType) *ChannelFilter {
	f.ChannelType = &channelType
	return f
}

// WithTags 設定標籤過濾
func (f *ChannelFilter) WithTags(tags []string) *ChannelFilter {
	f.Tags = tags
	return f
}

// WithEnabled 設定啟用狀態過濾
func (f *ChannelFilter) WithEnabled(enabled bool) *ChannelFilter {
	f.Enabled = &enabled
	return f
}

// HasChannelTypeFilter 檢查是否有通道類型過濾條件
func (f *ChannelFilter) HasChannelTypeFilter() bool {
	return f.ChannelType != nil
}

// HasTagsFilter 檢查是否有標籤過濾條件
func (f *ChannelFilter) HasTagsFilter() bool {
	return len(f.Tags) > 0
}

// HasEnabledFilter 檢查是否有啟用狀態過濾條件
func (f *ChannelFilter) HasEnabledFilter() bool {
	return f.Enabled != nil
}