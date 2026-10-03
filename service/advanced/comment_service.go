package advanced

import (
	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type CommentServiceGroup struct{}

func (c *CommentServiceGroup) CreateComment(ccrq *model.CreateCommentRequest) response.Code {
	co := ccrq.ToModel(ccrq.UserID)
	if co == nil {
		return response.CodeParamError
	}
	if co.ParentID != nil {
		rq := model.CommentsQuery{StartedID: co.ParentID, Limit: 1}
		oldco, err := dao.CommentDao.GetComments(&rq)
		if err != nil || oldco == nil {
			return response.CodeDatabaseError
		}
		co.RootID = oldco[0].RootID
	} else {
		co.RootID = 0
	}
	err := dao.CommentDao.CreateComment(co)
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// func (c *CommentServiceGroup) UpdateComment(ucrq *model.UpdateCommentRequest) response.Code {
// 	[]is_child := true
// 	rq := model.CommentsQuery{
// 		UserID: ucrq.ID,
// 	}
// 	co, err := dao.CommentDao.GetComments(&rq)
// 	if err != nil {
// 		return response.CodeCommentNotFound
// 	}
// 	rq = model.CommentsQuery{
// 		RootID: &co[0].RootID,
// 	}
// 	childs, err := dao.CommentDao.GetComments(&rq)
// }

// func (id int64) isChild() bool {
// 	return false
// }
