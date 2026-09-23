package dao

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"gorm.io/gorm"
)

type ItemImageGroup struct{}

// GetImagesByItemID 查询物品全部图片（按展示顺序升序）
func (itemImageGroup *ItemImageGroup) GetImagesByItemID(itemID int64) (images []model.ItemImage) {
	global.LNF_DB.Where("item_id = ?", itemID).
		Order("sort_order ASC").
		Find(&images)
	return images
}

// ReplaceImages 覆盖式设置物品图片：先删除旧图再批量插入（事务）
func (itemImageGroup *ItemImageGroup) ReplaceImages(itemID int64, inputs []model.ItemImageInput) error {
	return global.LNF_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("item_id = ?", itemID).Delete(&model.ItemImage{}).Error; err != nil {
			return err
		}
		for _, input := range inputs {
			img := &model.ItemImage{
				ItemID:    itemID,
				ImageURL:  input.ImageURL,
				SortOrder: input.SortOrder,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := tx.Create(img).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
