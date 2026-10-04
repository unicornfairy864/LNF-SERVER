package advanced

import (
	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type CommentServiceGroup struct{}

// 创建评论
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

// 管理员更新评论状态
func (c *CommentServiceGroup) UpdateComment(ucrq *model.UpdateCommentRequest) response.Code {
	qu := model.CommentsQuery{
		UserID: ucrq.ID,
	}
	co, err := dao.CommentDao.GetComments(&qu)
	if err != nil {
		return response.CodeCommentNotFound
	}
	qu = model.CommentsQuery{
		RootID: &co[0].RootID,
	}
	cos, err := dao.CommentDao.GetComments(&qu)
	is_child := make(map[int64]*bool, len(cos))
	mcos := make(map[int64]*model.Comment, len(cos))
	for _, item := range cos {
		mcos[item.ID] = item
	}
	for _, item := range mcos {
		isChild(mcos, is_child, nil, item.ID, *ucrq.ID)
	}
	ids := []*int64{}
	for id, isChild := range is_child {
		if *isChild {
			ids = append(ids, &id)
		}
	}
	err = dao.CommentDao.UpdateComment(ids, *ucrq.Status)
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// 获取批量列表
func (c *CommentServiceGroup) GetList(cgrq *model.CommentGetlistRequestQuery, iid *model.ItemRequestParam) (*model.CommentsListDTO, response.Code) {
	qu := model.CommentsQuery{
		ItemID:    &iid.ItemID,
		StartedID: &cgrq.StartedID,
		Limit:     cgrq.Limit,
	}
	cos, err := dao.CommentDao.GetComments(&qu)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	coDTOs := []*model.CommentDTO{}
	for _, item := range cos {
		coDTOs = append(coDTOs, item.ToDTO())
	}
	rt := model.ToList(len(cos), coDTOs)
	return rt, response.CodeSuccess
}

// 获取子评论
func (c *CommentServiceGroup) GetChildComments(gcrq *model.GetChildrenRequestQuery, iid *model.ItemRequestParam) (*model.CommentsListDTO, response.Code) {
	qu := model.CommentsQuery{
		ItemID:    &iid.ItemID,
		StartedID: &gcrq.ID,
		Limit:     1,
	}
	rco, err := dao.CommentDao.GetComments(&qu)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	qu = model.CommentsQuery{
		RootID: &rco[0].RootID,
	}
	cos, err := dao.CommentDao.GetComments(&qu)
	is_child := make(map[int64]*bool, len(cos))
	depth := make(map[int64]*int64, len(cos))
	mcos := make(map[int64]*model.Comment, len(cos))
	for _, item := range cos {
		mcos[item.ID] = item
	}
	for _, item := range cos {
		isChild(mcos, is_child, depth, item.ID, gcrq.ID)
	}
	de := gcrq.Depth
	for {
		count := int64(0)
		for id, d := range depth {
			if *is_child[id] && *d <= de {
				count++
			}
		}
		if count <= gcrq.MaxCount {
			break
		}
		if de <= 0 {
			return nil, response.CodeParamError
		}
		de--
	}
	ids := []int64{}
	for id, ic := range is_child {
		if *ic == true && *depth[id] <= de {
			ids = append(ids, id)
		}
	}
	qu = model.CommentsQuery{
		IDs: &ids,
	}
	cos, err = dao.CommentDao.GetComments(&qu)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	var coDTOs []*model.CommentDTO
	for _, item := range cos {
		coDTOs = append(coDTOs, item.ToDTO())
	}
	rt := model.ToList(len(cos), coDTOs)
	return rt, response.CodeSuccess
}

func isChild(mcos map[int64]*model.Comment, is_child map[int64]*bool, depth map[int64]*int64, id int64, rootID int64) (int64, bool) {
	ic, ok := is_child[id]
	nodp := false
	de := int64(-1)
	if depth == nil {
		nodp = true
	}
	if ok == true {
		if !nodp {
			de = *(depth[id])
		}
		return de, *ic
	}

	if id == rootID {
		*is_child[id] = true
		ic = is_child[id]
	}

	if mcos[id].RootID != 0 {
		pde, pic := isChild(mcos, is_child, depth, *mcos[id].ParentID, rootID)
		if !nodp {
			*depth[id] = pde + 1
			de = *depth[id]
		}
		if *ic == false {
			*ic = pic
		}
	} else {
		if !nodp {
			*depth[id] = 0
			de = *depth[id]
		}
	}
	return de, *ic
}
