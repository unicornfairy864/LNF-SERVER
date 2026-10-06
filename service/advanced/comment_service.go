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
	if co.ParentID != 0 {
		rq := model.CommentsQuery{
			ID: &ccrq.ParentID,
		}
		oldco, err := dao.CommentDao.GetComments(&rq)
		if err != nil {
			return response.CodeDatabaseError
		}
		if len(oldco) == 0 {
			return response.CodeCommentNotFound
		}
		co.RootID = oldco[0].RootID
	}

	err := dao.CommentDao.CreateComment(co)
	if err != nil {
		return response.CodeDatabaseError
	}
	if co.ParentID == 0 {
		co.RootID = co.ID
		err = dao.CommentDao.InnerUpdate(co)
	}
	if err != nil {
		dao.CommentDao.InnerDelete(co.ID)
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// 管理员更新评论状态
func (c *CommentServiceGroup) UpdateComment(ucrq *model.UpdateCommentRequest) response.Code {
	qu := model.CommentsQuery{
		ID: ucrq.ID,
	}
	co, err := dao.CommentDao.GetComments(&qu)
	if err != nil {
		return response.CodeDatabaseError
	}
	if co == nil {
		return response.CodeCommentNotFound
	}
	qu = model.CommentsQuery{
		RootID: &co[0].RootID,
	}
	cos, err := dao.CommentDao.GetComments(&qu)
	if err != nil {
		return response.CodeDatabaseError
	}
	is_child := make(map[int64]bool, len(cos))
	mcos := make(map[int64]*model.Comment, len(cos))
	for _, item := range cos {
		mcos[item.ID] = item
	}
	for _, item := range mcos {
		isChild(mcos, is_child, nil, item.ID, *ucrq.ID)
	}
	ids := []*int64{}
	for id, isChild := range is_child {
		if isChild {
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
func (c *CommentServiceGroup) GetList(cgrq *model.CommentGetlistRequestQuery, iid *model.CommentGetListRequestParam) (*model.CommentsListDTO, response.Code) {
	itemcheck := dao.ItemDao.GetItemByID(iid.ItemID)
	if itemcheck.ID == 0 {
		return nil, response.CodeItemNotFound
	}
	loop := int64(0)
	coDTOs := []*model.CommentDTO{}
	for {
		startat := cgrq.StartedID - loop*cgrq.Limit
		qu := model.CommentsQuery{
			ItemID:    &iid.ItemID,
			StartedID: &startat,
			Limit:     cgrq.Limit,
		}
		if startat <= 0 {
			rt := model.ToList(len(coDTOs), coDTOs)
			return rt, response.CodeSuccess
		}
		cos, err := dao.CommentDao.GetComments(&qu)
		if err != nil {
			return nil, response.CodeDatabaseError
		}
		for _, item := range cos {
			if item.Status == 1 {
				coDTOs = append(coDTOs, item.ToDTO())
			}
			if len(coDTOs) >= int(cgrq.Limit) {
				rt := model.ToList(len(coDTOs), coDTOs)
				return rt, response.CodeSuccess
			}
		}
		loop++
	}
}

// 获取子评论
func (c *CommentServiceGroup) GetChildComments(gcrq *model.GetChildrenRequest) (*model.CommentsListDTO, response.Code) {
	if gcrq.MaxCount == 0 {
		gcrq.MaxCount = 100
	}
	qu := model.CommentsQuery{
		ID: &gcrq.ID,
	}
	rco, err := dao.CommentDao.GetComments(&qu)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	if rco == nil {
		return nil, response.CodeCommentNotFound
	}
	qu = model.CommentsQuery{
		RootID: &rco[0].RootID,
	}
	cos, err := dao.CommentDao.GetComments(&qu)
	is_child := make(map[int64]bool, len(cos))
	depth := make(map[int64]int64, len(cos))
	mcos := make(map[int64]*model.Comment, len(cos))
	for _, item := range cos {
		if item.Status == 1 {
			mcos[item.ID] = item
		}
	}
	for _, item := range mcos {
		isChild(mcos, is_child, depth, item.ID, gcrq.ID)
	}
	de := gcrq.Depth + depth[gcrq.ID]
	var demax int64
	for _, item := range depth {
		if demax < item {
			demax = item
		}
	}
	if de > demax {
		de = demax
	}
	for {
		count := int64(0)
		for id, d := range depth {
			if is_child[id] && d <= de {
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
		if ic == true && depth[id] <= de && depth[id] > depth[gcrq.ID] {
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

func isChild(mcos map[int64]*model.Comment, is_child map[int64]bool, depth map[int64]int64, id int64, rootID int64) (int64, bool) {
	ic, ok := is_child[id]
	nodp := false
	de := int64(-1)
	if depth == nil {
		nodp = true
	}
	if ok == true {
		if !nodp {
			de = depth[id]
		}
		return de, ic
	}

	if id == rootID {
		is_child[id] = true
		ic = is_child[id]
	}

	if mcos[id].RootID != mcos[id].ID {
		pde, pic := isChild(mcos, is_child, depth, mcos[id].ParentID, rootID)
		if !nodp {
			depth[id] = pde + 1
			de = depth[id]
		}
		if ic == false {
			is_child[id] = pic
			ic = pic
		}
	} else {
		if !nodp {
			depth[id] = 0
			de = depth[id]
		}
	}
	return de, ic
}
