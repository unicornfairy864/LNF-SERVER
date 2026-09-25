package advanced

import (
	"strings"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type LocationServiceGroup struct{}

const (
	maxLocationNameLen = 100
	rootLocationLevel  = 1
	maxLocationDepth   = 50
)

// ListService 查询地点列表（parentID / level 均可选）
func (locationService *LocationServiceGroup) ListService(req *model.ListLocationRequest) ([]model.Location, response.Code) {
	return dao.LocationDao.GetLocations(req.ParentID, req.Level), response.CodeSuccess
}

// GetItemLocationsService 查询物品地点链（根→叶）
func (locationService *LocationServiceGroup) GetItemLocationsService(itemID int64) ([]model.Location, response.Code) {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return make([]model.Location, 0), response.CodeItemNotFound
	}
	chain := make([]model.Location, 0)
	if item.LocationID != nil && *item.LocationID > 0 {
		chain = dao.LocationDao.GetLocationChain(*item.LocationID)
	}
	return chain, response.CodeSuccess
}

// CreateService 创建地点（level 按父级计算：parent_id=0 为根节点 level=1）
func (locationService *LocationServiceGroup) CreateService(req *model.CreateLocationRequest) response.Code {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > maxLocationNameLen {
		return response.CodeParamError
	}
	level := rootLocationLevel
	// 指针字段仅在显式传入时生效，未传（nil）按默认值处理：parent_id=0 视为根节点，sort_order=0
	parentID := int64(0)
	if req.ParentID != nil {
		parentID = *req.ParentID
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	if parentID > 0 {
		parent := dao.LocationDao.GetLocationByID(parentID)
		if parent.ID == 0 {
			return response.CodeLocationNotFound
		}
		level = parent.Level + 1
	}
	// 同父级下名称查重
	if exists := dao.LocationDao.GetLocationByParentAndName(parentID, name); exists.ID != 0 {
		return response.CodeLocationDuplicate
	}
	location := &model.Location{
		Name:      name,
		ParentID:  parentID,
		Level:     level,
		Address:   req.Address,
		SortOrder: sortOrder,
	}
	if err := dao.LocationDao.CreateLocation(location); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// UpdateService 增量更新地点（非整体替换）；改父级时校验存在/非自身/非后代，并重算子树 level
func (locationService *LocationServiceGroup) UpdateService(req *model.UpdateLocationRequest) response.Code {
	location := dao.LocationDao.GetLocationByID(req.ID)
	if location.ID == 0 {
		return response.CodeLocationNotFound
	}
	updates := make(map[string]interface{})
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > maxLocationNameLen {
			return response.CodeParamError
		}
		updates["name"] = name
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	// 改父级：单独走 MoveLocation（事务中一并写入其他字段）
	if req.ParentID != nil && *req.ParentID != location.ParentID {
		newParentID := *req.ParentID
		if newParentID == req.ID {
			return response.CodeParamError
		}
		newParentLevel := 0
		if newParentID > 0 {
			newParent := dao.LocationDao.GetLocationByID(newParentID)
			if newParent.ID == 0 {
				return response.CodeLocationNotFound
			}
			// 防环：新父级不能是自身后代
			if isDescendant(newParentID, req.ID) {
				return response.CodeParamError
			}
			newParentLevel = newParent.Level
		}
		if err := dao.LocationDao.MoveLocation(req.ID, newParentID, newParentLevel, updates); err != nil {
			return response.CodeDatabaseError
		}
		return response.CodeSuccess
	}
	if len(updates) == 0 {
		return response.CodeSuccess
	}
	if err := dao.LocationDao.UpdateLocationByVK(req.ID, updates); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// isDescendant 判断 candidate 是否为 rootID 的后代（沿 parent 上溯，限深防环）
func isDescendant(candidateID int64, rootID int64) bool {
	cur := candidateID
	for i := 0; i < maxLocationDepth && cur > 0; i++ {
		if cur == rootID {
			return true
		}
		loc := dao.LocationDao.GetLocationByID(cur)
		if loc.ID == 0 {
			return false
		}
		cur = loc.ParentID
	}
	return false
}

// DeleteService 硬删除地点：先查子元素，再查物品引用
func (locationService *LocationServiceGroup) DeleteService(locationID int64) response.Code {
	location := dao.LocationDao.GetLocationByID(locationID)
	if location.ID == 0 {
		return response.CodeLocationNotFound
	}
	if dao.LocationDao.CountChildren(locationID) > 0 {
		return response.CodeLocationHasChildren
	}
	if dao.ItemDao.CountItemsByLocationID(locationID) > 0 {
		return response.CodeLocationInUse
	}
	if err := dao.LocationDao.DeleteLocationByID(locationID); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}
