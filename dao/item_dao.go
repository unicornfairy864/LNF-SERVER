package dao

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	modeladv "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"gorm.io/gorm"
)

type ItemGroup struct{}

// buildItemQuery 组装物品分页查询条件（每次调用返回全新 DB，避免 Count 污染链）
func buildItemQuery(q *model.ListItemQuery, userID int64, defaultStatus *int8) *gorm.DB {
	db := global.LNF_DB.Model(&model.Item{}).Where("is_deleted = 0")
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if q.Type != nil {
		db = db.Where("type = ?", *q.Type)
	}
	status := q.Status
	if status == nil {
		status = defaultStatus
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if q.LocationID != nil {
		db = db.Where("location_id = ?", *q.LocationID)
	}
	if q.TagID != nil {
		db = db.Where("id IN (SELECT item_id FROM item_tags WHERE tag_id = ?)", *q.TagID)
	}
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		db = db.Where("(title LIKE ? OR description LIKE ?)", kw, kw)
	}
	return db
}

// GetItemPage 分页查询物品；userID>0 时只查该用户；defaultStatus 仅在 q.Status 为空时生效
func (itemGroup *ItemGroup) GetItemPage(q *model.ListItemQuery, userID int64, defaultStatus *int8) (items []model.Item, total int64, err error) {
	var count int64
	if err = buildItemQuery(q, userID, defaultStatus).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	offset := (q.Page - 1) * q.PageSize
	err = buildItemQuery(q, userID, defaultStatus).
		Order("created_at DESC, id DESC").
		Offset(offset).Limit(q.PageSize).
		Find(&items).Error
	return items, count, err
}

// GetItemByID 按 ID 查询未删除物品
func (itemGroup *ItemGroup) GetItemByID(id int64) (item model.Item) {
	global.LNF_DB.Where(ConditionIDNotDeleted, id).First(&item)
	return item
}

// CreateItemWithTags 创建物品并写入标签关联（事务）
func (itemGroup *ItemGroup) CreateItemWithTags(item *model.Item, tagIDs []int64) error {
	return global.LNF_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		if len(tagIDs) == 0 {
			return nil
		}
		for _, tagID := range tagIDs {
			it := &modeladv.ItemTag{
				ItemID:    item.ID,
				TagID:     tagID,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(it).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateItemWithTags 增量更新物品；tagIDs 非 nil 时整体替换标签关联（事务）
func (itemGroup *ItemGroup) UpdateItemWithTags(id int64, updates map[string]interface{}, tagIDs *[]int64) error {
	return global.LNF_DB.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			updates["updated_at"] = time.Now()
			if err := tx.Model(&model.Item{}).
				Where(ConditionIDNotDeleted, id).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		if tagIDs != nil {
			if err := tx.Where("item_id = ?", id).Delete(&modeladv.ItemTag{}).Error; err != nil {
				return err
			}
			seen := make(map[int64]struct{}, len(*tagIDs))
			for _, tagID := range *tagIDs {
				if _, ok := seen[tagID]; ok {
					continue
				}
				seen[tagID] = struct{}{}
				it := &modeladv.ItemTag{
					ItemID:    id,
					TagID:     tagID,
					CreatedAt: time.Now(),
				}
				if err := tx.Create(it).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// SoftDeleteItem 软删除物品（仅本人）
func (itemGroup *ItemGroup) SoftDeleteItem(id int64, userID int64) error {
	return global.LNF_DB.Model(&model.Item{}).
		Where("id = ? AND user_id = ? AND is_deleted = 0", id, userID).
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"updated_at": time.Now(),
		}).Error
}

// IncrViewCount 浏览量 +1
func (itemGroup *ItemGroup) IncrViewCount(id int64) error {
	return global.LNF_DB.Model(&model.Item{}).
		Where("id = ? AND is_deleted = 0", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// CountItemsByLocationID 统计引用某地点的未删除物品数
func (itemGroup *ItemGroup) CountItemsByLocationID(locationID int64) (count int64) {
	global.LNF_DB.Model(&model.Item{}).
		Where("location_id = ? AND is_deleted = 0", locationID).
		Count(&count)
	return count
}
