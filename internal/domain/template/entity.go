package template

import (
	"errors"

	"channel-api/internal/domain/shared"
)

// Template 範本聚合根
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

// NewTemplate 建立新範本
func NewTemplate(
	name *TemplateName,
	description *Description,
	channelType shared.ChannelType,
	subject *Subject,
	content *TemplateContent,
	tags *Tags,
) (*Template, error) {
	// 驗證必要欄位
	if name == nil {
		return nil, errors.New("template name is required")
	}
	if !channelType.IsValid() {
		return nil, errors.New("invalid channel type")
	}
	if content == nil {
		return nil, errors.New("template content is required")
	}

	// 設定預設值
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

// ReconstructTemplate 從持久化資料重建範本
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

// ID 取得範本 ID
func (t *Template) ID() *TemplateID {
	return t.id
}

// Name 取得範本名稱
func (t *Template) Name() *TemplateName {
	return t.name
}

// Description 取得描述
func (t *Template) Description() *Description {
	return t.description
}

// ChannelType 取得通道類型
func (t *Template) ChannelType() shared.ChannelType {
	return t.channelType
}

// Subject 取得主題
func (t *Template) Subject() *Subject {
	return t.subject
}

// Content 取得範本內容
func (t *Template) Content() *TemplateContent {
	return t.content
}

// Tags 取得標籤
func (t *Template) Tags() *Tags {
	return t.tags
}

// Timestamps 取得時間戳記
func (t *Template) Timestamps() *shared.Timestamps {
	return t.timestamps
}

// Version 取得版本號
func (t *Template) Version() *Version {
	return t.version
}

// Update 更新範本
func (t *Template) Update(
	name *TemplateName,
	description *Description,
	channelType shared.ChannelType,
	subject *Subject,
	content *TemplateContent,
	tags *Tags,
) error {
	// 驗證必要欄位
	if name == nil {
		return errors.New("template name is required")
	}
	if !channelType.IsValid() {
		return errors.New("invalid channel type")
	}
	if content == nil {
		return errors.New("template content is required")
	}

	// 設定預設值
	if description == nil {
		description, _ = NewDescription("")
	}
	if subject == nil {
		subject, _ = NewSubject("")
	}
	if tags == nil {
		tags = NewTags(nil)
	}

	// 更新欄位
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

// Delete 軟刪除範本
func (t *Template) Delete() error {
	if t.timestamps.IsDeleted() {
		return errors.New("template is already deleted")
	}
	t.timestamps.MarkDeleted()
	return nil
}

// IsDeleted 檢查範本是否已刪除
func (t *Template) IsDeleted() bool {
	return t.timestamps.IsDeleted()
}

// HasTag 檢查是否包含指定標籤
func (t *Template) HasTag(tag string) bool {
	return t.tags.Contains(tag)
}

// HasAnyTag 檢查是否包含任一指定標籤
func (t *Template) HasAnyTag(tags []string) bool {
	return t.tags.ContainsAny(tags)
}

// MatchesType 檢查通道類型是否匹配
func (t *Template) MatchesType(channelType shared.ChannelType) bool {
	return t.channelType == channelType
}

// GetAllVariables 取得範本中的所有變數
func (t *Template) GetAllVariables() []string {
	variables := make(map[string]bool)
	
	// 從主題中提取變數
	for _, variable := range t.subject.ExtractVariables() {
		variables[variable] = true
	}
	
	// 從內容中提取變數
	for _, variable := range t.content.ExtractVariables() {
		variables[variable] = true
	}
	
	// 轉換為切片
	result := make([]string, 0, len(variables))
	for variable := range variables {
		result = append(result, variable)
	}
	
	return result
}

// ValidateVariables 驗證提供的變數是否包含範本所需的所有變數
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