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

func (c *CommentGroup) GetComments(cq *model.CommentsQuery) ([]*model.Comment, error) {
	var cos []*model.Comment
	var limit int
	db := global.LNF_DB

	if cq.ItemID != nil {
		db = db.Where("item_id = ?", cq.ItemID)
	}
	if cq.UserID != nil {
		db = db.Where("user_id = ?", cq.UserID)
	}
	if cq.RootID != nil {
		db = db.Where("root_id = ?", cq.RootID)
	}
	if cq.StartedID != nil && *cq.StartedID > 0 {
		db = db.Where("id <= ?", cq.StartedID)
	}
	if cq.IDs != nil {
		db = db.Where("id IN ?", *cq.IDs)
	}
	if cq.Limit != 0 {
		limit = int(cq.Limit)
	}
	db = db.Order("id DESC").Limit(limit).Find(&cos)

	err := db.Error
	if err != nil {
		return nil, err
	}
	return cos, nil
}
