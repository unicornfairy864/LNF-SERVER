package dao

import (
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
)

type ItemTagGroup struct{}

// GetTagsByItemID 联查物品关联的全部标签
func (itemTagGroup *ItemTagGroup) GetTagsByItemID(itemID int64) (tags []model.Tag) {
	global.LNF_DB.
		Table("tags").
		Select("tags.*").
		Joins("JOIN item_tags ON item_tags.tag_id = tags.id").
		Where("item_tags.item_id = ?", itemID).
		Order("tags.sort_order ASC, tags.id ASC").
		Find(&tags)
	return tags
}

// CountByTagID 统计标签被物品引用的次数
func (itemTagGroup *ItemTagGroup) CountByTagID(tagID int64) (count int64) {
	global.LNF_DB.Table("item_tags").
		Where("tag_id = ?", tagID).
		Count(&count)
	return count
}
