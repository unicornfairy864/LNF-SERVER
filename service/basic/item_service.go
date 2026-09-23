package basic

import (
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	modeladv "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type ItemServiceGroup struct{}

const (
	itemDefaultPage     = 1
	itemDefaultPageSize = 10
	itemMyMaxPageSize   = 50
	itemMaxTitleLen     = 100
	itemMaxImages       = 3
)

// normalizePage 归一化分页参数
func normalizePage(q *model.ListItemQuery) {
	if q.Page <= 0 {
		q.Page = itemDefaultPage
	}
	if q.PageSize <= 0 {
		q.PageSize = itemDefaultPageSize
	}
}

// buildResponse 组装物品响应（地点链 + 图片 + 标签）
func buildResponse(item *model.Item) *model.ItemResponse {
	resp := &model.ItemResponse{
		ID:             item.ID,
		UserID:         item.UserID,
		Title:          item.Title,
		Description:    item.Description,
		Type:           item.Type,
		Status:         item.Status,
		LocationDetail: item.LocationDetail,
		Images:         make([]model.ItemImage, 0),
		Tags:           make([]modeladv.Tag, 0),
		LostFoundTime:  item.LostFoundTime,
		Contact:        item.Contact,
		CreditReward:   item.CreditReward,
		ViewCount:      item.ViewCount,
		ClaimUserID:    item.ClaimUserID,
		ClaimTime:      item.ClaimTime,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
	resp.Locations = make([]modeladv.Location, 0)
	if item.LocationID != nil && *item.LocationID > 0 {
		resp.Locations = dao.LocationDao.GetLocationChain(*item.LocationID)
	}
	resp.Images = dao.ItemImageDao.GetImagesByItemID(item.ID)
	resp.Tags = dao.ItemTagDao.GetTagsByItemID(item.ID)
	return resp
}

// validateTagIDs 校验标签全部存在（去重后），返回去重后的 ID 列表
func validateTagIDs(tagIDs []int64) ([]int64, response.Code) {
	seen := make(map[int64]struct{}, len(tagIDs))
	dedup := make([]int64, 0, len(tagIDs))
	for _, id := range tagIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		dedup = append(dedup, id)
	}
	if len(dedup) == 0 {
		return dedup, response.CodeSuccess
	}
	tags := dao.TagDao.GetTagsByIDs(dedup)
	if len(tags) != len(dedup) {
		return nil, response.CodeTagNotFound
	}
	return dedup, response.CodeSuccess
}

// ListPublicService 公开物品列表（未指定 status 时默认只看已发布 status=0）
func (itemService *ItemServiceGroup) ListPublicService(q *model.ListItemQuery) (*model.ItemListResponse, response.Code) {
	normalizePage(q)
	status := int8(0)
	items, total, err := dao.ItemDao.GetItemPage(q, 0, &status)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return pageResponse(q, total, items), response.CodeSuccess
}

// ListMyService 我的发布列表（默认不过滤状态；分页默认10，page_size 上限50）
func (itemService *ItemServiceGroup) ListMyService(userID int64, q *model.ListItemQuery) (*model.ItemListResponse, response.Code) {
	normalizePage(q)
	if q.PageSize > itemMyMaxPageSize {
		return nil, response.CodeParamError
	}
	items, total, err := dao.ItemDao.GetItemPage(q, userID, nil)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return pageResponse(q, total, items), response.CodeSuccess
}

// pageResponse 组装分页响应
func pageResponse(q *model.ListItemQuery, total int64, items []model.Item) *model.ItemListResponse {
	responses := make([]model.ItemResponse, 0, len(items))
	for i := range items {
		responses = append(responses, *buildResponse(&items[i]))
	}
	return &model.ItemListResponse{
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
		Items:    responses,
	}
}

// GetDetailService 物品详情（浏览量 +1）
func (itemService *ItemServiceGroup) GetDetailService(itemID int64) (*model.ItemResponse, response.Code) {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return nil, response.CodeItemNotFound
	}
	// 浏览量异步失败不影响详情返回
	_ = dao.ItemDao.IncrViewCount(itemID)
	item.ViewCount++
	return buildResponse(&item), response.CodeSuccess
}

// CreateService 创建物品（status 默认 0 已发布）
func (itemService *ItemServiceGroup) CreateService(userID int64, req *model.CreateItemRequest) response.Code {
	if req.Type != 0 && req.Type != 1 {
		return response.CodeItemTypeInvalid
	}
	if strings.TrimSpace(req.Title) == "" || len(req.Title) > itemMaxTitleLen {
		return response.CodeParamError
	}
	// 地点校验
	if req.LocationID != nil {
		if dao.LocationDao.GetLocationByID(*req.LocationID).ID == 0 {
			return response.CodeItemLocationInvalid
		}
	}
	// 标签校验
	tagIDs, code := validateTagIDs(req.TagIDs)
	if code != response.CodeSuccess {
		return code
	}
	locationID := req.LocationID
	item := &model.Item{
		UserID:         userID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           req.Type,
		Status:         0,
		LocationID:     locationID,
		LocationDetail: req.LocationDetail,
		LostFoundTime:  req.LostFoundTime,
		Contact:        req.Contact,
		CreditReward:   int64(req.CreditReward),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := dao.ItemDao.CreateItemWithTags(item, tagIDs); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// UpdateService 增量更新物品（仅发布者本人）
func (itemService *ItemServiceGroup) UpdateService(userID int64, req *model.UpdateItemRequest) response.Code {
	item := dao.ItemDao.GetItemByID(req.ID)
	if item.ID == 0 {
		return response.CodeItemNotFound
	}
	if item.UserID != userID {
		return response.CodeItemNoPermission
	}
	updates := make(map[string]interface{})
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" || len(*req.Title) > itemMaxTitleLen {
			return response.CodeParamError
		}
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		if *req.Status < 0 || *req.Status > 2 {
			return response.CodeParamError
		}
		updates["status"] = *req.Status
	}
	if req.LocationID != nil {
		if *req.LocationID > 0 && dao.LocationDao.GetLocationByID(*req.LocationID).ID == 0 {
			return response.CodeItemLocationInvalid
		}
		updates["location_id"] = *req.LocationID
	}
	if req.LocationDetail != nil {
		updates["location_detail"] = *req.LocationDetail
	}
	if req.LostFoundTime != nil {
		updates["lost_found_time"] = *req.LostFoundTime
	}
	if req.Contact != nil {
		updates["contact"] = *req.Contact
	}
	if req.CreditReward != nil {
		if *req.CreditReward < 0 {
			return response.CodeParamError
		}
		updates["credit_reward"] = *req.CreditReward
	}
	if req.ClaimUserID != nil {
		updates["claim_user_id"] = *req.ClaimUserID
	}
	if req.ClaimTime != nil {
		updates["claim_time"] = *req.ClaimTime
	}
	// 标签整体替换
	var tagIDs *[]int64
	if req.TagIDs != nil {
		dedup, code := validateTagIDs(*req.TagIDs)
		if code != response.CodeSuccess {
			return code
		}
		tagIDs = &dedup
	}
	if len(updates) == 0 && tagIDs == nil {
		return response.CodeSuccess
	}
	if err := dao.ItemDao.UpdateItemWithTags(req.ID, updates, tagIDs); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// DeleteService 软删除物品（仅发布者本人）
func (itemService *ItemServiceGroup) DeleteService(userID int64, itemID int64) response.Code {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return response.CodeItemNotFound
	}
	if item.UserID != userID {
		return response.CodeItemNoPermission
	}
	if err := dao.ItemDao.SoftDeleteItem(itemID, userID); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// SetImagesService 覆盖式设置物品图片（仅发布者本人）
func (itemService *ItemServiceGroup) SetImagesService(userID int64, itemID int64, req *model.SetItemImagesRequest) response.Code {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return response.CodeItemNotFound
	}
	if item.UserID != userID {
		return response.CodeItemNoPermission
	}
	if len(req.Images) > itemMaxImages {
		return response.CodeItemImageTooMany
	}
	seen := make(map[int8]struct{}, len(req.Images))
	for _, img := range req.Images {
		if img.SortOrder < 1 || img.SortOrder > itemMaxImages {
			return response.CodeItemImageInvalid
		}
		if _, ok := seen[img.SortOrder]; ok {
			return response.CodeItemImageInvalid
		}
		seen[img.SortOrder] = struct{}{}
		if strings.TrimSpace(img.ImageURL) == "" || len(img.ImageURL) > 500 {
			return response.CodeItemImageInvalid
		}
	}
	if err := dao.ItemImageDao.ReplaceImages(itemID, req.Images); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}
