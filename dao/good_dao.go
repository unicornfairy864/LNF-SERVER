package dao

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"gorm.io/gorm"
)

type GoodGroup struct{}

// GetGoodByID 按 ID 查询未删除商品
func (goodGroup *GoodGroup) GetGoodByID(id int64) (good model.Good) {
	global.LNF_DB.Where(ConditionIDNotDeleted, id).First(&good)
	return good
}

// GetGoodByName 按名称查询商品（创建/改名查重用；仅看未删除，下架商品名可复用）
func (goodGroup *GoodGroup) GetGoodByName(name string) (good model.Good) {
	global.LNF_DB.Where("name = ? AND is_deleted = 0", name).First(&good)
	return good
}

// buildGoodQuery 组装商品分页查询条件（每次调用返回全新 DB，避免 Count 污染链）；
// 含 keyword 名称模糊匹配与 min_price/max_price 积分闭区间筛选（0 = 该侧不限，
// 价格条件走 idx_price(is_deleted,price,sort_order) 复合索引）
func buildGoodQuery(q *model.ListGoodQuery) *gorm.DB {
	db := global.LNF_DB.Model(&model.Good{}).Where("is_deleted = 0")
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		db = db.Where("name LIKE ?", kw)
	}
	if q.MinPrice > 0 {
		db = db.Where("price >= ?", q.MinPrice)
	}
	if q.MaxPrice > 0 {
		db = db.Where("price <= ?", q.MaxPrice)
	}
	return db
}

// GetGoodPage 分页查询商品（公开列表：仅未删除；keyword/积分区间筛选见 buildGoodQuery；sort_order 升序，其次创建时间降序）
func (goodGroup *GoodGroup) GetGoodPage(q *model.ListGoodQuery) (goods []model.Good, total int64, err error) {
	var count int64
	if err = buildGoodQuery(q).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	offset := (q.Page - 1) * q.PageSize
	err = buildGoodQuery(q).
		Order("sort_order ASC, created_at DESC, id DESC").
		Offset(offset).Limit(q.PageSize).
		Find(&goods).Error
	return goods, count, err
}

// CreateGood 创建商品
func (goodGroup *GoodGroup) CreateGood(good *model.Good) error {
	return global.LNF_DB.Create(good).Error
}

// UpdateGoodByVK 按 ID 增量更新商品（仅未删除商品）
func (goodGroup *GoodGroup) UpdateGoodByVK(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return global.LNF_DB.Model(&model.Good{}).
		Where(ConditionIDNotDeleted, id).
		Updates(updates).Error
}

// SoftDeleteGood 软删除（下架）商品
func (goodGroup *GoodGroup) SoftDeleteGood(id int64) error {
	return global.LNF_DB.Model(&model.Good{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"updated_at": time.Now(),
		}).Error
}

// RedeemStockTx 在指定事务上条件更新扣减库存：stock>0 才扣，防并发超卖；
// affected=0 表示库存不足（或商品已被下架/删除）
func (goodGroup *GoodGroup) RedeemStockTx(tx *gorm.DB, goodID int64) (int64, error) {
	res := tx.Model(&model.Good{}).
		Where("id = ? AND is_deleted = 0 AND stock > 0", goodID).
		UpdateColumn("stock", gorm.Expr("stock - 1"))
	return res.RowsAffected, res.Error
}
