package dao

import (
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"gorm.io/gorm"
)

type OrderGroup struct{}

// CreateOrderTx 在指定事务上写入兑换订单（快照字段由 service 填充）
func (orderGroup *OrderGroup) CreateOrderTx(tx *gorm.DB, order *model.Order) error {
	return tx.Create(order).Error
}

// GetOrdersByUserID 分页查询用户兑换记录（created_at 降序）
func (orderGroup *OrderGroup) GetOrdersByUserID(userID int64, page int, pageSize int) (orders []model.Order, total int64, err error) {
	var count int64
	if err = global.LNF_DB.Model(&model.Order{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err = global.LNF_DB.
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error
	return orders, count, err
}
