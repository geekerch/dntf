package template

import (
	"context"

	"channel-api/internal/domain/shared"
)

// TemplateRepository 範本倉儲介面
type TemplateRepository interface {
	// Save 儲存範本
	Save(ctx context.Context, template *Template) error
	
	// FindByID 根據 ID 查找範本
	FindByID(ctx context.Context, id *TemplateID) (*Template, error)
	
	// FindByName 根據名稱查找範本
	FindByName(ctx context.Context, name *TemplateName) (*Template, error)
	
	// FindAll 查找所有範本 (支援分頁和過濾)
	FindAll(ctx context.Context, filter *TemplateFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*Template], error)
	
	// Update 更新範本
	Update(ctx context.Context, template *Template) error
	
	// Delete 刪除範本
	Delete(ctx context.Context, id *TemplateID) error
	
	// Exists 檢查範本是否存在
	Exists(ctx context.Context, id *TemplateID) (bool, error)
	
	// ExistsByName 檢查指定名稱的範本是否存在
	ExistsByName(ctx context.Context, name *TemplateName) (bool, error)
}

// TemplateFilter 範本過濾條件
type TemplateFilter struct {
	ChannelType *shared.ChannelType `json:"channelType,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
}

// NewTemplateFilter 建立範本過濾條件
func NewTemplateFilter() *TemplateFilter {
	return &TemplateFilter{}
}

// WithChannelType 設定通道類型過濾
func (f *TemplateFilter) WithChannelType(channelType shared.ChannelType) *TemplateFilter {
	f.ChannelType = &channelType
	return f
}

// WithTags 設定標籤過濾
func (f *TemplateFilter) WithTags(tags []string) *TemplateFilter {
	f.Tags = tags
	return f
}

// HasChannelTypeFilter 檢查是否有通道類型過濾條件
func (f *TemplateFilter) HasChannelTypeFilter() bool {
	return f.ChannelType != nil
}

// HasTagsFilter 檢查是否有標籤過濾條件
func (f *TemplateFilter) HasTagsFilter() bool {
	return len(f.Tags) > 0
}