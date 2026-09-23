package dao

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"gorm.io/gorm"
)

type LocationGroup struct{}

// GetLocationByID 按 ID 查询地点
func (locationGroup *LocationGroup) GetLocationByID(id int64) (location model.Location) {
	global.LNF_DB.Where("id = ?", id).First(&location)
	return location
}

// GetLocationByParentAndName 按父地点+名称查询（查重用）
func (locationGroup *LocationGroup) GetLocationByParentAndName(parentID int64, name string) (location model.Location) {
	global.LNF_DB.Where("parent_id = ? AND name = ?", parentID, name).First(&location)
	return location
}

// GetLocations 查询地点列表（parentID / level 为 nil 时不按该条件过滤）
func (locationGroup *LocationGroup) GetLocations(parentID *int64, level *int) (locations []model.Location) {
	db := global.LNF_DB.Model(&model.Location{})
	if parentID != nil {
		db = db.Where("parent_id = ?", *parentID)
	}
	if level != nil {
		db = db.Where("level = ?", *level)
	}
	db.Order("sort_order ASC, id ASC").Find(&locations)
	return locations
}

// GetLocationsByIDs 批量按 ID 查询地点
func (locationGroup *LocationGroup) GetLocationsByIDs(ids []int64) (locations []model.Location) {
	if len(ids) == 0 {
		return locations
	}
	global.LNF_DB.Where("id IN (?)", ids).Find(&locations)
	return locations
}

// CreateLocation 创建地点
func (locationGroup *LocationGroup) CreateLocation(location *model.Location) error {
	return global.LNF_DB.Create(location).Error
}

// UpdateLocationByVK 按 ID 增量更新地点字段
func (locationGroup *LocationGroup) UpdateLocationByVK(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return global.LNF_DB.Model(&model.Location{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// CountChildren 统计某地点的直接子元素数量
func (locationGroup *LocationGroup) CountChildren(parentID int64) (count int64) {
	global.LNF_DB.Model(&model.Location{}).
		Where("parent_id = ?", parentID).
		Count(&count)
	return count
}

// GetLocationChain 沿 parent 上溯获取地点链（返回 根→叶 顺序；含自身；限深 50 防环）
func (locationGroup *LocationGroup) GetLocationChain(locationID int64) (chain []model.Location) {
	chain = make([]model.Location, 0)
	cur := locationID
	for i := 0; i < 50 && cur > 0; i++ {
		loc := locationGroup.GetLocationByID(cur)
		if loc.ID == 0 {
			break
		}
		chain = append(chain, loc)
		cur = loc.ParentID
	}
	// 反转为 根→叶
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

// MoveLocation 移动地点到新父级：更新自身 parent_id/level，并在事务中自上而下重算整棵子树 level
// extraUpdates 为同一事务中一并写入的其他字段更新（name/address/sort_order 等），可为 nil
func (locationGroup *LocationGroup) MoveLocation(id int64, newParentID int64, newParentLevel int, extraUpdates map[string]interface{}) error {
	return global.LNF_DB.Transaction(func(tx *gorm.DB) error {
		newLevel := newParentLevel + 1
		updates := map[string]interface{}{
			"parent_id":  newParentID,
			"level":      newLevel,
			"updated_at": time.Now(),
		}
		for k, v := range extraUpdates {
			updates[k] = v
		}
		if err := tx.Model(&model.Location{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return err
		}
		// BFS 重算子树 level
		frontier := []int64{id}
		curLevel := newLevel
		for len(frontier) > 0 {
			var children []model.Location
			if err := tx.Where("parent_id IN ?", frontier).Find(&children).Error; err != nil {
				return err
			}
			if len(children) == 0 {
				break
			}
			curLevel++
			childIDs := make([]int64, 0, len(children))
			for _, c := range children {
				childIDs = append(childIDs, c.ID)
			}
			if err := tx.Model(&model.Location{}).
				Where("id IN ?", childIDs).
				Updates(map[string]interface{}{
					"level":      curLevel,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
			frontier = childIDs
		}
		return nil
	})
}

// DeleteLocationByID 硬删除地点（locations 表无 is_deleted；引用/子元素检查在 service 完成）
func (locationGroup *LocationGroup) DeleteLocationByID(id int64) error {
	return global.LNF_DB.Where("id = ?", id).Delete(&model.Location{}).Error
}
