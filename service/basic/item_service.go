package basic

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	"github.com/unicornfairy864/LNF-SERVER/global"
	modeladv "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service/advanced"
)

type ItemServiceGroup struct{}

const (
	// 长度上限对齐 items.sql 列定义（title varchar(100)、location_detail varchar(200)、contact varchar(100)），
	// 均按 UTF-8 字节数判定：字节数 ≤ 上限 ⇒ 字符数 ≤ 上限，不会超出列宽
	itemDefaultPage          = 1
	itemDefaultPageSize      = 10
	itemMaxPageSize          = 50
	itemMyMaxPageSize        = 50
	itemMaxTitleLen          = 100
	itemMaxLocationDetailLen = 200
	itemMaxContactLen        = 100
	itemMaxImages            = 3

	// 物品状态（items.status）
	itemStatusPublished int8 = 0 // 已发布
	itemStatusClaimed   int8 = 1 // 已认领
	itemStatusClosed    int8 = 2 // 已关闭

	// 认领自动关闭单批扫描上限
	itemAutoCloseBatchSize = 200

	// 检索时 location 计分只检测该层级的地点（非 level3 的传入 ID 自动忽略且不计入条件总数）
	itemSearchLocationLevel = 3

	// 积分流水类型（见 credit_logs.sql type 注释）：1 认领成功奖励
	creditLogTypeClaimReward int64 = 1

	// 通知类型（见 notifications.sql type 注释）：3 认领结果、5 积分变动
	notificationTypeClaimResult  int8 = 3
	notificationTypeCreditChange int8 = 5
)

// listDefaultStatuses 列表接口未显式指定 status 时的默认状态范围：已发布(0) + 已认领(1)
var listDefaultStatuses = []int8{itemStatusPublished, itemStatusClaimed}

// CountSearchingService 统计正在被寻找的物品数量（已发布0 + 已认领1 的未删除物品）
func (itemService *ItemServiceGroup) CountSearchingService() (int64, response.Code) {
	count, err := dao.ItemDao.CountSearchingItems(listDefaultStatuses)
	if err != nil {
		return 0, response.CodeDatabaseError
	}
	return count, response.CodeSuccess
}

// notificationService 通知服务实例（service/basic 包内共享：item 认领关闭/确认 + user QQ 绑定）；
// 直接依赖 service/advanced 而非 service 包单例，避免 basic ↔ service 循环引用
var notificationService = &advanced.NotificationServiceGroup{}

// normalizePage 归一化分页参数
func normalizePage(q *model.ListItemQuery) {
	if q.Page <= 0 {
		q.Page = itemDefaultPage
	}
	if q.PageSize <= 0 || q.PageSize > itemMaxPageSize {
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

// ListPublicService 公开物品列表（未指定 status 时默认返回已发布与已认领 status=0/1）
func (itemService *ItemServiceGroup) ListPublicService(q *model.ListItemQuery) (*model.ItemListResponse, response.Code) {
	normalizePage(q)
	items, total, err := dao.ItemDao.GetItemPage(q, 0, listDefaultStatuses)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return pageResponse(q, total, items), response.CodeSuccess
}

// ListMyService 我的发布列表（未指定 status 时默认返回已发布与已认领；分页默认10，page_size 上限50）
func (itemService *ItemServiceGroup) ListMyService(userID int64, q *model.ListItemQuery) (*model.ItemListResponse, response.Code) {
	normalizePage(q)
	if q.PageSize > itemMyMaxPageSize {
		return nil, response.CodeParamError
	}
	items, total, err := dao.ItemDao.GetItemPage(q, userID, listDefaultStatuses)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return pageResponse(q, total, items), response.CodeSuccess
}

// pageResponse 组装分页响应
func pageResponse(q *model.ListItemQuery, total int64, items []model.Item) *model.ItemListResponse {
	return buildPagedList(q.Page, q.PageSize, total, items)
}

// buildPagedList 组装分页响应（页码/页大小显式传入，供列表与检索共用）
func buildPagedList(page int, pageSize int, total int64, items []model.Item) *model.ItemListResponse {
	responses := make([]model.ItemResponse, 0, len(items))
	for i := range items {
		responses = append(responses, *buildResponse(&items[i]))
	}
	return &model.ItemListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    responses,
	}
}

// SearchService 多条件最小匹配检索（tag + location 计分，返回 match_count >= min_match 的物品）
// 计分：match_count = |item.tags ∩ tag_ids| + (item.location_id ∈ location_ids(仅level3) ? 1 : 0)
// 校验：min_match 必传且 ≥1；status 必传可多选（仅 0/1，禁 2）；
// 条件总数 = 去重后真实存在的 tag 数 + level3 地点数（"剔除计"：无法得分的条件不计入），
// 要求 min_match ≤ 条件总数（相等 = 全部条件必须满足），两组均为空时条件总数为 0 同样被拦截
func (itemService *ItemServiceGroup) SearchService(q *model.ItemMatchQuery) (*model.ItemListResponse, response.Code) {
	// min_match ≥1（binding 已保证，防御性复核）
	if q.MinMatch < 1 {
		return nil, response.CodeParamError
	}
	// status 必传、多选、仅允许 0/1（binding 已保证，此处去重+防御）
	statuses := make([]int8, 0, len(q.Status))
	seenStatus := make(map[int8]struct{}, len(q.Status))
	for _, s := range q.Status {
		if s != itemStatusPublished && s != itemStatusClaimed {
			return nil, response.CodeParamError
		}
		if _, ok := seenStatus[s]; ok {
			continue
		}
		seenStatus[s] = struct{}{}
		statuses = append(statuses, s)
	}
	if len(statuses) == 0 {
		return nil, response.CodeParamError
	}
	// tag 去重后仅保留真实存在的（与 location "剔除计"口径一致：无法得分的条件不计入条件总数）
	seenTag := make(map[int64]struct{}, len(q.TagIDs))
	dedupTags := make([]int64, 0, len(q.TagIDs))
	for _, id := range q.TagIDs {
		if _, ok := seenTag[id]; ok {
			continue
		}
		seenTag[id] = struct{}{}
		dedupTags = append(dedupTags, id)
	}
	tagIDs := make([]int64, 0, len(dedupTags))
	for _, tag := range dao.TagDao.GetTagsByIDs(dedupTags) {
		tagIDs = append(tagIDs, tag.ID)
	}
	// location 仅保留 level=3（非 level3 / 不存在的 ID 自动忽略，且不计入条件总数）
	locationIDs := make([]int64, 0, len(q.LocationIDs))
	for _, loc := range dao.LocationDao.GetLocationsByIDs(q.LocationIDs) {
		if loc.Level == itemSearchLocationLevel {
			locationIDs = append(locationIDs, loc.ID)
		}
	}
	// 条件总数与 min_match 校验
	if q.MinMatch > len(tagIDs)+len(locationIDs) {
		return nil, response.CodeParamError
	}
	if q.Page <= 0 {
		q.Page = itemDefaultPage
	}
	if q.PageSize <= 0 || q.PageSize > itemMaxPageSize {
		q.PageSize = itemDefaultPageSize
	}
	items, total, err := dao.ItemDao.GetItemsByMinMatch(q, tagIDs, locationIDs, q.MinMatch, statuses)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return buildPagedList(q.Page, q.PageSize, total, items), response.CodeSuccess
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
	if *req.Type != 0 && *req.Type != 1 {
		return response.CodeItemTypeInvalid
	}
	if strings.TrimSpace(req.Title) == "" || len(req.Title) > itemMaxTitleLen {
		return response.CodeParamError
	}
	// 可选字段校验：location_detail ≤200、contact ≤100（字节），credit_reward ≥0
	if req.LocationDetail != nil && len(*req.LocationDetail) > itemMaxLocationDetailLen {
		return response.CodeParamError
	}
	if req.Contact != nil && len(*req.Contact) > itemMaxContactLen {
		return response.CodeParamError
	}
	if req.CreditReward < 0 {
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
		Type:           *req.Type,
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
	// status 状态流转由专门接口管理（claim/withdraw/confirm/close），不允许通过 update 修改
	if req.LocationID != nil {
		if *req.LocationID > 0 && dao.LocationDao.GetLocationByID(*req.LocationID).ID == 0 {
			return response.CodeItemLocationInvalid
		}
		updates["location_id"] = *req.LocationID
	}
	if req.LocationDetail != nil {
		if len(*req.LocationDetail) > itemMaxLocationDetailLen {
			return response.CodeParamError
		}
		updates["location_detail"] = *req.LocationDetail
	}
	if req.LostFoundTime != nil {
		updates["lost_found_time"] = *req.LostFoundTime
	}
	if req.Contact != nil {
		if len(*req.Contact) > itemMaxContactLen {
			return response.CodeParamError
		}
		updates["contact"] = *req.Contact
	}
	if req.CreditReward != nil {
		if *req.CreditReward < 0 {
			return response.CodeParamError
		}
		updates["credit_reward"] = *req.CreditReward
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

// ==================== 认领 / 关闭 ====================

// claimBeneficiary 认领成功积分受益人：拾物帖(type=1)归发帖者，失物帖(type=0)归认领者
func claimBeneficiary(item *model.Item) int64 {
	if item.Type == 1 {
		return item.UserID
	}
	if item.ClaimUserID != nil {
		return *item.ClaimUserID
	}
	return 0
}

// notifyClaimClosed 认领关闭结果通知（type=3 认领结果，系统触发 adminID=0，relatedID=物品ID）。
// auto 区分超时自动关闭与发布者手动确认；isBeneficiary 为接收者是否积分受益人
// （拾物帖受益人=发帖者，失物帖受益人=认领者）；发送失败仅记日志，不影响关闭/发分主流程
func notifyClaimClosed(receiverID int64, item *model.Item, credit int64, isBeneficiary bool, auto bool) {
	if receiverID == 0 {
		return
	}
	title, verb := "认领已确认", "已被发布者确认"
	if auto {
		title, verb = "认领已超时自动确认", "因认领超时已自动确认"
	}
	role := "你认领的物品"
	if receiverID == item.UserID {
		role = "你发布的物品"
	}
	content := fmt.Sprintf("%s「%s」%s，物品已关闭", role, item.Title, verb)
	if isBeneficiary {
		content += fmt.Sprintf("，拾金不昧积分奖励 %d 分已发放至你的账户。", credit)
	} else {
		content += "。"
	}
	relatedID := item.ID
	if err := notificationService.Create(0, receiverID, notificationTypeClaimResult, title, content, &relatedID); err != nil {
		log.Printf("[item] 认领关闭通知发送失败 item_id=%d receiver_id=%d: %v", item.ID, receiverID, err)
	}
}

// notifySelfCloseToClaimer 发帖者自行找回关闭物品时，通知仍在认领中的认领者
// （type=3 认领结果，系统触发 adminID=0，relatedID=物品ID；该路径不发积分，
// 与 notifyClaimClosed 的确认/超时语义区分）；发送失败仅记日志，不影响关闭主流程
func notifySelfCloseToClaimer(claimUserID int64, item *model.Item) {
	if claimUserID == 0 {
		return
	}
	content := fmt.Sprintf("你认领的物品「%s」已被发布者自行找回并关闭，本次认领结束。", item.Title)
	relatedID := item.ID
	if err := notificationService.Create(0, claimUserID, notificationTypeClaimResult,
		"认领已结束", content, &relatedID); err != nil {
		log.Printf("[item] 自行关闭认领中止通知发送失败 item_id=%d claim_user_id=%d: %v", item.ID, claimUserID, err)
	}
}

// notifyClaimApplied 有人认领物品时通知发帖人（type=3 认领结果，系统触发 adminID=0，relatedID=物品ID；
// 发送失败仅记日志，不影响认领主流程）
func notifyClaimApplied(posterID int64, item *model.Item) {
	if posterID == 0 {
		return
	}
	relatedID := item.ID
	content := fmt.Sprintf("你发布的物品「%s」已被其他同学认领，请尽快核实；确认找回请在“我的发布”中确认，若信息不符可撤回该认领。", item.Title)
	if err := notificationService.Create(0, posterID, notificationTypeClaimResult, "有人认领了你的物品", content, &relatedID); err != nil {
		log.Printf("[item] 认领申请通知发送失败 item_id=%d poster_id=%d: %v", item.ID, posterID, err)
	}
}

// notifyClaimWithdrawn 撤回认领后通知另一方（type=3 认领结果，系统触发 adminID=0，relatedID=物品ID）。
// actorID 为撤回者：认领者撤回 → 通知发帖人；发帖者撤回 → 通知认领者；发送失败仅记日志
func notifyClaimWithdrawn(item *model.Item, actorID int64, claimUserID int64) {
	receiverID := item.UserID
	receiverIsPoster := true
	if actorID != claimUserID {
		// 撤回者是发帖者 → 通知认领者
		receiverID, receiverIsPoster = claimUserID, false
	}
	if receiverID == 0 || receiverID == actorID {
		return
	}
	role := "你认领的物品"
	if receiverIsPoster {
		role = "你发布的物品"
	}
	relatedID := item.ID
	content := fmt.Sprintf("%s「%s」的认领已被撤回，物品恢复为“已发布”状态，可再次被认领。", role, item.Title)
	if err := notificationService.Create(0, receiverID, notificationTypeClaimResult, "认领已撤回", content, &relatedID); err != nil {
		log.Printf("[item] 认领撤回通知发送失败 item_id=%d receiver_id=%d: %v", item.ID, receiverID, err)
	}
}

// claimRemindMarkTTL 超时提醒去重 key 存活时间（每个物品只提醒一次）
const claimRemindMarkTTL = 24 * time.Hour

// notifyCreditReward 认领奖励积分变动通知（type=5 积分变动，系统触发 adminID=0，relatedID=物品ID）。
// 与 type=3 的认领结果通知并存：前者叙述事件，本条记录积分变动（与商城兑换的双通知风格一致）；
// 发送失败仅记日志，不影响关闭/发分主流程
func notifyCreditReward(beneficiaryID int64, credit int64, item *model.Item, auto bool) {
	if beneficiaryID == 0 || credit == 0 {
		return
	}
	reason := "认领成功奖励"
	if auto {
		reason = "认领超时自动关闭奖励"
	}
	balance := dao.UserDao.GetUserByID(beneficiaryID).Credit
	relatedID := item.ID
	content := fmt.Sprintf("你的积分增加 %d 分（%s），当前余额 %d 分。", credit, reason, balance)
	if err := notificationService.Create(0, beneficiaryID, notificationTypeCreditChange, "积分变动", content, &relatedID); err != nil {
		log.Printf("[item] 积分变动通知发送失败 item_id=%d user_id=%d: %v", item.ID, beneficiaryID, err)
	}
}

// notifyClaimTimeoutReminder 认领即将超时提醒发帖人（type=3 认领结果，系统触发 adminID=0）
func notifyClaimTimeoutReminder(item *model.Item, remindBefore time.Duration) {
	if item.UserID == 0 {
		return
	}
	relatedID := item.ID
	content := fmt.Sprintf("你发布的物品「%s」已被认领；若 %s 内未处理，系统将按“确认由他人找回”自动关闭并发分，请尽快在“我的发布”中确认或撤回认领。",
		item.Title, humanDuration(remindBefore))
	if err := notificationService.Create(0, item.UserID, notificationTypeClaimResult, "认领即将超时", content, &relatedID); err != nil {
		log.Printf("[item] 认领超时提醒发送失败 item_id=%d poster_id=%d: %v", item.ID, item.UserID, err)
	}
}

// humanDuration 时长中文描述（仅用于通知文案）
func humanDuration(d time.Duration) string {
	switch {
	case d >= 24*time.Hour && d%(24*time.Hour) == 0:
		return fmt.Sprintf("%d 天", int(d/(24*time.Hour)))
	case d >= time.Hour && d%time.Hour == 0:
		return fmt.Sprintf("%d 小时", int(d/time.Hour))
	case d >= time.Minute && d%time.Minute == 0:
		return fmt.Sprintf("%d 分钟", int(d/time.Minute))
	default:
		return d.String()
	}
}

// claimRemindMarkOnce 同一物品只提醒一次（Redis INCR 原子；Redis 异常时放行，宁可重复提醒也不漏）
func claimRemindMarkOnce(itemID int64) bool {
	key := "lnf:notify:claim-remind:" + strconv.FormatInt(itemID, 10)
	n, err := dao.RedisDao.INCR(key)
	if err != nil {
		return true
	}
	if n == 1 {
		_ = dao.RedisDao.SetKey(key, n, claimRemindMarkTTL)
	}
	return n == 1
}

// RemindExpiringClaimsService 认领即将超时时提醒发帖人（在自动关闭前 remindBefore 窗口内，每物品仅一次）
// 由 initialization.StartClaimAutoCloseScheduler 定时调用（先提醒、后关闭）
func (itemService *ItemServiceGroup) RemindExpiringClaimsService() {
	duration := global.LNF_CONFIG.Server.ClaimAutoClose
	remindBefore := global.LNF_CONFIG.Server.ClaimRemindBeforeValue()
	if duration <= 0 || remindBefore <= 0 || remindBefore >= duration {
		return
	}
	now := time.Now()
	from := now.Add(-duration)   // 已到期时刻（不含）
	to := from.Add(remindBefore) // 提醒窗口上界（含）
	items, err := dao.ItemDao.ListClaimsApproachingDeadline(from, to, itemAutoCloseBatchSize)
	if err != nil {
		return
	}
	for i := range items {
		item := &items[i]
		if item.ClaimUserID == nil || *item.ClaimUserID == item.UserID {
			continue
		}
		if !claimRemindMarkOnce(item.ID) {
			continue
		}
		notifyClaimTimeoutReminder(item, remindBefore)
	}
}

// ClaimService 认领物品（登录用户；是否要求绑定QQ由 server.claim_qq_required 控制）
func (itemService *ItemServiceGroup) ClaimService(userID int64, itemID int64) response.Code {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return response.CodeItemNotFound
	}
	if item.Status == itemStatusClosed {
		return response.CodeItemClosed
	}
	if item.Status != itemStatusPublished {
		return response.CodeItemAlreadyClaimed
	}
	if item.UserID == userID {
		return response.CodeClaimSelfItem
	}
	user := dao.UserDao.GetUserByID(userID)
	if user.ID == 0 || user.Status == 0 {
		return response.CodeUserNotFoundOrBanned
	}
	if global.LNF_CONFIG.Server.ClaimQQRequired && (user.QQ == nil || strings.TrimSpace(*user.QQ) == "") {
		return response.CodeClaimQQRequired
	}
	affected, err := dao.ItemDao.ClaimItem(itemID, userID, time.Now())
	if err != nil {
		return response.CodeDatabaseError
	}
	if affected == 0 {
		// 并发下状态已变化，重查给出准确错误
		item = dao.ItemDao.GetItemByID(itemID)
		if item.Status == itemStatusClosed {
			return response.CodeItemClosed
		}
		return response.CodeItemAlreadyClaimed
	}
	// 通知发帖人：有人认领了其发布的物品
	notifyClaimApplied(item.UserID, &item)
	return response.CodeSuccess
}

// WithdrawClaimService 撤回认领（认领者或发帖者双方均可，仅 status=1 未关闭时可撤；撤回后恢复为已发布）
func (itemService *ItemServiceGroup) WithdrawClaimService(userID int64, itemID int64) response.Code {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return response.CodeItemNotFound
	}
	if item.Status == itemStatusClosed {
		return response.CodeItemClosed
	}
	if item.Status != itemStatusClaimed || item.ClaimUserID == nil {
		return response.CodeClaimNotFound
	}
	if userID != *item.ClaimUserID && userID != item.UserID {
		return response.CodeClaimNoPermission
	}
	affected, err := dao.ItemDao.WithdrawClaim(itemID)
	if err != nil {
		return response.CodeDatabaseError
	}
	if affected == 0 {
		item = dao.ItemDao.GetItemByID(itemID)
		if item.Status == itemStatusClosed {
			return response.CodeItemClosed
		}
		return response.CodeClaimNotFound
	}
	// 通知另一方（认领者撤回→发帖人；发帖者撤回→认领者）
	notifyClaimWithdrawn(&item, userID, *item.ClaimUserID)
	return response.CodeSuccess
}

// ConfirmClaimService 发帖者确认由他人找回：关闭物品并给拾到者加 server.claim_credit 积分
func (itemService *ItemServiceGroup) ConfirmClaimService(userID int64, itemID int64) response.Code {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return response.CodeItemNotFound
	}
	if item.UserID != userID {
		return response.CodeItemNoPermission
	}
	if item.Status == itemStatusClosed {
		return response.CodeItemClosed
	}
	if item.Status != itemStatusClaimed || item.ClaimUserID == nil {
		return response.CodeClaimNotFound
	}
	beneficiary := claimBeneficiary(&item)
	if beneficiary == 0 {
		return response.CodeServerError
	}
	affected, err := dao.ItemDao.CloseItemWithCredit(itemID, beneficiary,
		global.LNF_CONFIG.Server.ClaimCredit, creditLogTypeClaimReward, "认领成功奖励")
	if err != nil {
		return response.CodeDatabaseError
	}
	if affected == 0 {
		item = dao.ItemDao.GetItemByID(itemID)
		if item.Status == itemStatusClosed {
			return response.CodeItemClosed
		}
		return response.CodeClaimNotFound
	}
	// 事务提交成功：通知认领者（确认者=发帖者本人，不自我通知）；
	// 失物帖(type=0)认领者为积分受益人，文案带积分发放说明
	notifyClaimClosed(*item.ClaimUserID, &item, global.LNF_CONFIG.Server.ClaimCredit,
		beneficiary == *item.ClaimUserID, false)
	// 积分变动通知（type=5）：与认领结果通知并存
	notifyCreditReward(beneficiary, global.LNF_CONFIG.Server.ClaimCredit, &item, false)
	return response.CodeSuccess
}

// CloseSelfService 发帖者关闭自己的帖子（自己已经找回，不发积分；status=0/1 均可关）
func (itemService *ItemServiceGroup) CloseSelfService(userID int64, itemID int64) response.Code {
	item := dao.ItemDao.GetItemByID(itemID)
	if item.ID == 0 {
		return response.CodeItemNotFound
	}
	if item.UserID != userID {
		return response.CodeItemNoPermission
	}
	if item.Status == itemStatusClosed {
		return response.CodeItemClosed
	}
	affected, err := dao.ItemDao.CloseItem(itemID)
	if err != nil {
		return response.CodeDatabaseError
	}
	if affected == 0 {
		return response.CodeItemClosed
	}
	// 关闭前存在进行中的认领（status=1）时，通知认领者其认领随关闭结束（该路径不发积分）；
	// item 为关闭前读取的内存副本，ClaimUserID 仍可用（CloseItem 已清空 DB 侧字段）
	if item.Status == itemStatusClaimed && item.ClaimUserID != nil && *item.ClaimUserID != item.UserID {
		notifySelfCloseToClaimer(*item.ClaimUserID, &item)
	}
	return response.CodeSuccess
}

// AutoCloseExpiredClaimsService 扫描认领超时的物品并自动关闭（按确认认领语义发分）
// 由 initialization.StartClaimAutoCloseScheduler 定时调用
func (itemService *ItemServiceGroup) AutoCloseExpiredClaimsService() {
	duration := global.LNF_CONFIG.Server.ClaimAutoClose
	if duration <= 0 {
		return
	}
	deadline := time.Now().Add(-duration)
	items, err := dao.ItemDao.ListExpiredClaimedItems(deadline, itemAutoCloseBatchSize)
	if err != nil {
		return
	}
	for i := range items {
		item := &items[i]
		beneficiary := claimBeneficiary(item)
		if beneficiary == 0 {
			continue
		}
		// 条件更新保证并发下只关闭/发分一次；affected==1 时通知发帖者与认领者双方
		affected, cerr := dao.ItemDao.CloseItemWithCredit(item.ID, beneficiary,
			global.LNF_CONFIG.Server.ClaimCredit, creditLogTypeClaimReward, "认领超时自动关闭奖励")
		if cerr != nil || affected == 0 {
			continue
		}
		notifyClaimClosed(item.UserID, item, global.LNF_CONFIG.Server.ClaimCredit,
			beneficiary == item.UserID, true)
		if item.ClaimUserID != nil {
			notifyClaimClosed(*item.ClaimUserID, item, global.LNF_CONFIG.Server.ClaimCredit,
				beneficiary == *item.ClaimUserID, true)
		}
		// 积分变动通知（type=5）
		notifyCreditReward(beneficiary, global.LNF_CONFIG.Server.ClaimCredit, item, true)
	}
}
