package template

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// TemplateID 範本唯一識別碼
type TemplateID struct {
	value string
}

// NewTemplateID 建立新的範本 ID
func NewTemplateID() *TemplateID {
	return &TemplateID{
		value: "tpl_" + uuid.New().String(),
	}
}

// NewTemplateIDFromString 從字串建立範本 ID
func NewTemplateIDFromString(id string) (*TemplateID, error) {
	if id == "" {
		return nil, errors.New("template ID cannot be empty")
	}
	return &TemplateID{value: id}, nil
}

// String 返回字串表示
func (t *TemplateID) String() string {
	return t.value
}

// Equals 比較兩個範本 ID 是否相等
func (t *TemplateID) Equals(other *TemplateID) bool {
	if other == nil {
		return false
	}
	return t.value == other.value
}

// TemplateName 範本名稱
type TemplateName struct {
	value string
}

// NewTemplateName 建立範本名稱
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

// String 返回字串表示
func (t *TemplateName) String() string {
	return t.value
}

// Equals 比較兩個範本名稱是否相等
func (t *TemplateName) Equals(other *TemplateName) bool {
	if other == nil {
		return false
	}
	return t.value == other.value
}

// Subject 主題
type Subject struct {
	value string
}

// NewSubject 建立主題
func NewSubject(subject string) (*Subject, error) {
	subject = strings.TrimSpace(subject)
	if len(subject) > 200 {
		return nil, errors.New("subject cannot exceed 200 characters")
	}
	return &Subject{value: subject}, nil
}

// String 返回字串表示
func (s *Subject) String() string {
	return s.value
}

// IsEmpty 檢查主題是否為空
func (s *Subject) IsEmpty() bool {
	return s.value == ""
}

// HasVariables 檢查主題是否包含變數
func (s *Subject) HasVariables() bool {
	return strings.Contains(s.value, "{") && strings.Contains(s.value, "}")
}

// ExtractVariables 提取主題中的變數
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

// TemplateContent 範本內容
type TemplateContent struct {
	value string
}

// NewTemplateContent 建立範本內容
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

// String 返回字串表示
func (t *TemplateContent) String() string {
	return t.value
}

// HasVariables 檢查內容是否包含變數
func (t *TemplateContent) HasVariables() bool {
	return strings.Contains(t.value, "{") && strings.Contains(t.value, "}")
}

// ExtractVariables 提取內容中的變數
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

// Version 版本號
type Version struct {
	value int
}

// NewVersion 建立版本號
func NewVersion() *Version {
	return &Version{value: 1}
}

// NewVersionFromInt 從整數建立版本號
func NewVersionFromInt(version int) (*Version, error) {
	if version < 1 {
		return nil, errors.New("version must be positive")
	}
	return &Version{value: version}, nil
}

// Int 返回整數表示
func (v *Version) Int() int {
	return v.value
}

// Increment 遞增版本號
func (v *Version) Increment() *Version {
	return &Version{value: v.value + 1}
}

// Equals 比較兩個版本號是否相等
func (v *Version) Equals(other *Version) bool {
	if other == nil {
		return false
	}
	return v.value == other.value
}

// IsGreaterThan 檢查是否大於另一個版本號
func (v *Version) IsGreaterThan(other *Version) bool {
	if other == nil {
		return true
	}
	return v.value > other.value
}

// Description 描述
type Description struct {
	value string
}

// NewDescription 建立描述
func NewDescription(desc string) (*Description, error) {
	desc = strings.TrimSpace(desc)
	if len(desc) > 500 {
		return nil, errors.New("description cannot exceed 500 characters")
	}
	return &Description{value: desc}, nil
}

// String 返回字串表示
func (d *Description) String() string {
	return d.value
}

// IsEmpty 檢查描述是否為空
func (d *Description) IsEmpty() bool {
	return d.value == ""
}

// Tags 標籤
type Tags struct {
	tags []string
}

// NewTags 建立標籤
func NewTags(tags []string) *Tags {
	if tags == nil {
		tags = make([]string, 0)
	}
	// 去重並排序
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

// Add 新增標籤
func (t *Tags) Add(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}
	// 檢查是否已存在
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return
		}
	}
	t.tags = append(t.tags, tag)
}

// Remove 移除標籤
func (t *Tags) Remove(tag string) {
	for i, existingTag := range t.tags {
		if existingTag == tag {
			t.tags = append(t.tags[:i], t.tags[i+1:]...)
			return
		}
	}
}

// Contains 檢查是否包含標籤
func (t *Tags) Contains(tag string) bool {
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// ContainsAny 檢查是否包含任一標籤
func (t *Tags) ContainsAny(tags []string) bool {
	for _, tag := range tags {
		if t.Contains(tag) {
			return true
		}
	}
	return false
}

// ToSlice 轉換為切片
func (t *Tags) ToSlice() []string {
	result := make([]string, len(t.tags))
	copy(result, t.tags)
	return result
}

// Count 取得標籤數量
func (t *Tags) Count() int {
	return len(t.tags)
}
