package dao

import (
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	modeladv "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"gorm.io/gorm"
)

type ItemGroup struct{}

// buildItemQuery 组装物品分页查询条件（每次调用返回全新 DB，避免 Count 污染链）
func buildItemQuery(q *model.ListItemQuery, userID int64, defaultStatuses []int8) *gorm.DB {
	db := global.LNF_DB.Model(&model.Item{}).Where("is_deleted = 0")
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if q.Type != nil {
		db = db.Where("type = ?", *q.Type)
	}
	if q.Status != nil {
		db = db.Where("status = ?", *q.Status)
	} else if len(defaultStatuses) > 0 {
		db = db.Where("status IN ?", defaultStatuses)
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

// buildMinMatchExpr 构造最小匹配计分表达式（match_count = 标签命中数 + 地点命中分）
// 与 service 约定：tagIDs/locationIDs 已去重且已剔除无法得分的 ID（tag 不存在、location 非 level3），
// 因此允许其中一个为空（此时该维度计 0 分），不允许两个都为空
func buildMinMatchExpr(tagIDs []int64, locationIDs []int64) (string, []interface{}) {
	exprs := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if len(tagIDs) > 0 {
		exprs = append(exprs,
			"(SELECT COUNT(*) FROM item_tags it WHERE it.item_id = items.id AND it.tag_id IN (?))")
		args = append(args, tagIDs)
	} else {
		exprs = append(exprs, "0")
	}
	if len(locationIDs) > 0 {
		exprs = append(exprs,
			"(CASE WHEN items.location_id IN (?) THEN 1 ELSE 0 END)")
		args = append(args, locationIDs)
	} else {
		exprs = append(exprs, "0")
	}
	return strings.Join(exprs, " + "), args
}

// GetItemsByMinMatch 多条件最小匹配分页查询：
// match_count = |item.tags ∩ tagIDs| + (item.location_id ∈ locationIDs ? 1 : 0) >= minMatch
// statuses 由 service 保证非空（仅 0/1）；排序 created_at DESC, id DESC
func (itemGroup *ItemGroup) GetItemsByMinMatch(q *model.ItemMatchQuery, tagIDs []int64, locationIDs []int64, minMatch int, statuses []int8) (items []model.Item, total int64, err error) {
	matchExpr, matchArgs := buildMinMatchExpr(tagIDs, locationIDs)
	base := func() *gorm.DB {
		db := global.LNF_DB.Table("items").
			Where("is_deleted = 0").
			Where("("+matchExpr+") >= ?", append(matchArgs, minMatch)...).
			Where("status IN ?", statuses)
		if q.Type != nil {
			db = db.Where("type = ?", *q.Type)
		}
		return db
	}
	var count int64
	if err = base().Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err = base().
		Order("created_at DESC, id DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&items).Error
	return items, count, err
}

// GetItemPage 分页查询物品；userID>0 时只查该用户；defaultStatuses 仅在 q.Status 为空时生效（状态 IN 查询）
func (itemGroup *ItemGroup) GetItemPage(q *model.ListItemQuery, userID int64, defaultStatuses []int8) (items []model.Item, total int64, err error) {
	var count int64
	if err = buildItemQuery(q, userID, defaultStatuses).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	offset := (q.Page - 1) * q.PageSize
	err = buildItemQuery(q, userID, defaultStatuses).
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

// ==================== 认领 / 关闭 ====================

// ClaimItem 认领物品（条件更新防并发：仅 status=0 时生效）
func (itemGroup *ItemGroup) ClaimItem(itemID int64, claimUserID int64, claimTime time.Time) (int64, error) {
	res := global.LNF_DB.Model(&model.Item{}).
		Where("id = ? AND status = 0 AND is_deleted = 0", itemID).
		Updates(map[string]interface{}{
			"status":        1,
			"claim_user_id": claimUserID,
			"claim_time":    claimTime,
			"updated_at":    time.Now(),
		})
	return res.RowsAffected, res.Error
}

// WithdrawClaim 撤回认领（仅 status=1 时生效；恢复为已发布并清空认领字段）
func (itemGroup *ItemGroup) WithdrawClaim(itemID int64) (int64, error) {
	res := global.LNF_DB.Model(&model.Item{}).
		Where("id = ? AND status = 1 AND is_deleted = 0", itemID).
		Updates(map[string]interface{}{
			"status":        0,
			"claim_user_id": nil,
			"claim_time":    nil,
			"updated_at":    time.Now(),
		})
	return res.RowsAffected, res.Error
}

// CloseItem 关闭物品（发帖者自己找回，不发积分；status=0/1 → 2，清空认领字段）
func (itemGroup *ItemGroup) CloseItem(itemID int64) (int64, error) {
	res := global.LNF_DB.Model(&model.Item{}).
		Where("id = ? AND status IN ? AND is_deleted = 0", itemID, []int8{0, 1}).
		Updates(map[string]interface{}{
			"status":        2,
			"claim_user_id": nil,
			"claim_time":    nil,
			"updated_at":    time.Now(),
		})
	return res.RowsAffected, res.Error
}

// CloseItemWithCredit 确认认领并关闭物品（status=1 → 2），同一事务内给受益人加积分；
// 条件更新保证并发下只关闭/发分一次，affected=0 表示状态已变化（已被并发确认/撤回）
func (itemGroup *ItemGroup) CloseItemWithCredit(itemID int64, beneficiaryID int64, credit int64, logType int64, desc string) (int64, error) {
	affected := int64(0)
	err := global.LNF_DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Item{}).
			Where("id = ? AND status = 1 AND is_deleted = 0", itemID).
			Updates(map[string]interface{}{
				"status":     2,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		affected = res.RowsAffected
		return UserDao.AddUserCreditTx(tx, beneficiaryID, credit, logType, desc, 0)
	})
	return affected, err
}

// ListExpiredClaimedItems 查询认领超时仍未关闭的物品
func (itemGroup *ItemGroup) ListExpiredClaimedItems(deadline time.Time, limit int) ([]model.Item, error) {
	var items []model.Item
	err := global.LNF_DB.
		Where("status = 1 AND is_deleted = 0 AND claim_time <= ?", deadline).
		Order("claim_time ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}
