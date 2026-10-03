package dao

import (
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
)

type CommentGroup struct{}

func (c *CommentGroup) CreateComment(comment *model.Comment) error {
	err := global.LNF_DB.Create(comment).Error
	return err
}

func (c *CommentGroup) UpdateComment(ids []*int64, status int8) error {
	err := global.LNF_DB.Where("id IN ?", ids).Select("status").Update("status", status).Error
	return err
}

func (c *CommentGroup) GetComments(gc *model.CommentsQuery) ([]*model.Comment, error) {
	var cos []*model.Comment
	var limit int
	db := global.LNF_DB

	if gc.ItemID != nil {
		db = db.Where("item_id = ?", gc.ItemID)
	}
	if gc.UserID != nil {
		db = db.Where("user_id = ?", gc.UserID)
	}
	if gc.RootID != nil {
		db = db.Where("root_id = ?", gc.RootID)
	}
	if gc.StartedID != nil && *gc.StartedID > 0 {
		db = db.Where("id <= ?", gc.StartedID)
	}
	if gc.Limit != 0 {
		limit = int(gc.Limit)
	}
	db = db.Order("id DESC").Limit(limit).Find(&cos)

	err := db.Error
	if err != nil {
		return nil, err
	}
	return cos, nil
}
