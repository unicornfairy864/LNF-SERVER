package advanced

import (
	"strings"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type TagServiceGroup struct{}

const maxTagNameLen = 50

// ListService 查询全部标签
func (tagService *TagServiceGroup) ListService() ([]model.Tag, response.Code) {
	return dao.TagDao.GetTagList(), response.CodeSuccess
}

// CreateService 创建标签（名称唯一）
func (tagService *TagServiceGroup) CreateService(req *model.CreateTagRequest) response.Code {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > maxTagNameLen {
		return response.CodeTagNameInvalid
	}
	if dao.TagDao.GetTagByName(name).ID != 0 {
		return response.CodeTagDuplicate
	}
	tag := &model.Tag{
		Name:      name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	}
	if err := dao.TagDao.CreateTag(tag); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// UpdateService 增量更新标签（非整体替换）
func (tagService *TagServiceGroup) UpdateService(req *model.UpdateTagRequest) response.Code {
	tag := dao.TagDao.GetTagByID(req.ID)
	if tag.ID == 0 {
		return response.CodeTagNotFound
	}
	updates := make(map[string]interface{})
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > maxTagNameLen {
			return response.CodeTagNameInvalid
		}
		// 查重（排除自身）
		if exists := dao.TagDao.GetTagByName(name); exists.ID != 0 && exists.ID != req.ID {
			return response.CodeTagDuplicate
		}
		updates["name"] = name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) == 0 {
		return response.CodeSuccess
	}
	if err := dao.TagDao.UpdateTagByVK(req.ID, updates); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// DeleteService 硬删除标签（被物品引用则拒绝）
func (tagService *TagServiceGroup) DeleteService(tagID int64) response.Code {
	tag := dao.TagDao.GetTagByID(tagID)
	if tag.ID == 0 {
		return response.CodeTagNotFound
	}
	if dao.ItemTagDao.CountByTagID(tagID) > 0 {
		return response.CodeTagInUse
	}
	if err := dao.TagDao.DeleteTagByID(tagID); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}
