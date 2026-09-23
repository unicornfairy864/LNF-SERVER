package dao

import (
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
)

type TagGroup struct{}

// GetTagList 查询全部标签（排序序号升序）
func (tagGroup *TagGroup) GetTagList() (tags []model.Tag) {
	global.LNF_DB.Order("sort_order ASC, id ASC").Find(&tags)
	return tags
}

// GetTagByID 按 ID 查询标签
func (tagGroup *TagGroup) GetTagByID(id int64) (tag model.Tag) {
	global.LNF_DB.Where("id = ?", id).First(&tag)
	return tag
}

// GetTagByName 按名称查询标签（uk_name 唯一）
func (tagGroup *TagGroup) GetTagByName(name string) (tag model.Tag) {
	global.LNF_DB.Where("name = ?", name).First(&tag)
	return tag
}

// GetTagsByIDs 批量按 ID 查询标签
func (tagGroup *TagGroup) GetTagsByIDs(ids []int64) (tags []model.Tag) {
	if len(ids) == 0 {
		return tags
	}
	global.LNF_DB.Where("id IN (?)", ids).Find(&tags)
	return tags
}

// CreateTag 创建标签
func (tagGroup *TagGroup) CreateTag(tag *model.Tag) error {
	return global.LNF_DB.Create(tag).Error
}

// UpdateTagByVK 按 ID 增量更新标签
func (tagGroup *TagGroup) UpdateTagByVK(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	return global.LNF_DB.Model(&model.Tag{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteTagByID 硬删除标签（tags 表无 is_deleted）
func (tagGroup *TagGroup) DeleteTagByID(id int64) error {
	return global.LNF_DB.Where("id = ?", id).Delete(&model.Tag{}).Error
}
