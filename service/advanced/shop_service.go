package advanced

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/chensong"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"gorm.io/gorm"
)

type ShopServiceGroup struct{}

const (
	shopDefaultPage     = 1
	shopDefaultPageSize = 10
	goodMaxNameLen      = 100
	goodMaxImageURLLen  = 500
	goodMaxPrice        = 1000000

	// 积分流水类型（见 credit_logs.sql type 注释）：4 积分兑换
	creditLogTypeShopRedeem int64 = 4

	// 通知类型（见 notifications.sql type 注释）：5 积分变动、6 商品兑换
	notificationTypeCreditChange int8 = 5
	notificationTypeGoodsRedeem  int8 = 6

	// QQ 号合法数值区间：与用户模块（handler/basic/user_handler.go）判定一致，5~11 位
	qqMinValue = 10000
	qqMaxValue = 99999999999

	// 业务单号 = 14 位时间戳 + 6 位随机数
	orderNoTimeLayout  = "20060102150405"
	orderNoRandomWidth = 6
)

// errGoodStockNotEnough 事务内库存不足哨兵（并发下另一请求先扣完最后一秒库存）
var errGoodStockNotEnough = errors.New("good stock not enough")

// creditNotEnoughMsg 与 dao/user_dao.go AddUserCreditTx 的余额不足错误文案保持一致；
// 该处返回 fmt.Errorf 非哨兵错误，为不改动既有用户模块代码，此处以消息匹配映射为 50001
const creditNotEnoughMsg = "credit not enough"

// notificationService 通知服务实例（本包内直接实例化使用，避免 advanced ↔ service 循环引用）
var notificationService = &NotificationServiceGroup{}

// validateQQFormat 校验 QQ 号格式：5~11 位纯数值（区间判定与用户模块一致）
func validateQQFormat(qq string) bool {
	if qq == "" {
		return false
	}
	for _, r := range qq {
		if r < '0' || r > '9' {
			return false
		}
	}
	v, err := strconv.ParseInt(qq, 10, 64)
	if err != nil {
		return false
	}
	return v >= qqMinValue && v <= qqMaxValue
}

// generateOrderNo 生成业务单号：14 位时间戳 + 6 位随机数（幂等展示用，uk_order_no 兑底）
func generateOrderNo() string {
	return time.Now().Format(orderNoTimeLayout) +
		fmt.Sprintf("%0*d", orderNoRandomWidth, rand.Intn(1000000))
}

// ListService 公开商品列表（仅未删除；keyword 模糊匹配名称；min_price/max_price 闭区间
// 筛选含边界、可单边，min>max 返回参数错误；sort_order 升序）
func (shopService *ShopServiceGroup) ListService(q *model.ListGoodQuery) (*model.GoodListResponse, response.Code) {
	// 积分区间校验（负数已由 binding 拦截）：双边都传且 min>max → 参数错误
	if q.MinPrice > 0 && q.MaxPrice > 0 && q.MinPrice > q.MaxPrice {
		return nil, response.CodeParamError
	}
	if q.Page <= 0 {
		q.Page = shopDefaultPage
	}
	if q.PageSize <= 0 {
		q.PageSize = shopDefaultPageSize
	}
	goods, total, err := dao.GoodDao.GetGoodPage(q)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return &model.GoodListResponse{
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
		Items:    goods,
	}, response.CodeSuccess
}

// GetDetailService 商品详情
func (shopService *ShopServiceGroup) GetDetailService(goodID int64) (*model.Good, response.Code) {
	good := dao.GoodDao.GetGoodByID(goodID)
	if good.ID == 0 {
		return nil, response.CodeGoodsNotFound
	}
	return &good, response.CodeSuccess
}

// CreateService 创建商品（管理员）
func (shopService *ShopServiceGroup) CreateService(req *model.CreateGoodRequest) response.Code {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > goodMaxNameLen {
		return response.CodeGoodsNameInvalid
	}
	if dao.GoodDao.GetGoodByName(name).ID != 0 {
		return response.CodeGoodsNameInvalid
	}
	// price 缺失/非法统一由 service 判定（不做 binding 必填，错误码更精确）
	if req.Price < 1 || req.Price > goodMaxPrice {
		return response.CodeGoodsPriceInvalid
	}
	if req.Stock != nil && *req.Stock < 0 {
		return response.CodeParamError
	}
	if req.ImageURL != nil && (strings.TrimSpace(*req.ImageURL) == "" || len(*req.ImageURL) > goodMaxImageURLLen) {
		return response.CodeParamError
	}
	good := &model.Good{
		Name:        name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		SortOrder:   0,
	}
	if req.Stock != nil {
		good.Stock = *req.Stock
	}
	if req.SortOrder != nil {
		good.SortOrder = *req.SortOrder
	}
	if err := dao.GoodDao.CreateGood(good); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// UpdateService 增量更新商品（管理员，非整体替换；软删商品返回 11001）
func (shopService *ShopServiceGroup) UpdateService(req *model.UpdateGoodRequest) response.Code {
	good := dao.GoodDao.GetGoodByID(req.ID)
	if good.ID == 0 {
		return response.CodeGoodsNotFound
	}
	updates := make(map[string]interface{})
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > goodMaxNameLen {
			return response.CodeGoodsNameInvalid
		}
		// 查重（排除自身）
		if exists := dao.GoodDao.GetGoodByName(name); exists.ID != 0 && exists.ID != req.ID {
			return response.CodeGoodsNameInvalid
		}
		updates["name"] = name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ImageURL != nil {
		if strings.TrimSpace(*req.ImageURL) == "" || len(*req.ImageURL) > goodMaxImageURLLen {
			return response.CodeParamError
		}
		updates["image_url"] = *req.ImageURL
	}
	if req.Price != nil {
		if *req.Price < 1 || *req.Price > goodMaxPrice {
			return response.CodeGoodsPriceInvalid
		}
		updates["price"] = *req.Price
	}
	if req.Stock != nil {
		if *req.Stock < 0 {
			return response.CodeParamError
		}
		updates["stock"] = *req.Stock
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) == 0 {
		return response.CodeSuccess
	}
	if err := dao.GoodDao.UpdateGoodByVK(req.ID, updates); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// DeleteService 软删除（下架）商品（管理员）；已有兑换记录不受影响（订单为快照设计）
func (shopService *ShopServiceGroup) DeleteService(goodID int64) response.Code {
	good := dao.GoodDao.GetGoodByID(goodID)
	if good.ID == 0 {
		return response.CodeGoodsNotFound
	}
	if err := dao.GoodDao.SoftDeleteGood(goodID); err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// ListMyOrdersService 我的兑换记录（分页，created_at 降序）
func (shopService *ShopServiceGroup) ListMyOrdersService(userID int64, q *model.OrderListQuery) (*model.OrderListResponse, response.Code) {
	if q.Page <= 0 {
		q.Page = shopDefaultPage
	}
	if q.PageSize <= 0 {
		q.PageSize = shopDefaultPageSize
	}
	orders, total, err := dao.OrderDao.GetOrdersByUserID(userID, q.Page, q.PageSize)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return &model.OrderListResponse{
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
		Orders:   orders,
	}, response.CodeSuccess
}

// RedeemGoodsService 积分兑换商品（登录用户；流程详见 agent.md/shop_agent.md §四）：
// 事务内「条件扣库存（stock>0 防超卖）+ 扣积分（行锁，不足回滚）+ 写订单快照」；
// 事务提交成功后同步发两条站内通知（type=5 积分变动、type=6 商品兑换，见 notifyRedeemNotifications），
// 再异步发 chensong 群消息（均失败不影响兑换结果）
func (shopService *ShopServiceGroup) RedeemGoodsService(userID int64, goodID int64) (*model.RedeemGoodsResponse, response.Code) {
	// 1. 用户校验：存在、未禁用、已绑 QQ 且格式合法（5~11 位，见 §五）
	user := dao.UserDao.GetUserByID(userID)
	if user.ID == 0 || user.Status == 0 {
		return nil, response.CodeUserNotFoundOrBanned
	}
	if user.QQ == nil || !validateQQFormat(*user.QQ) {
		return nil, response.CodeShopQQRequired
	}
	// 2. 商品校验
	good := dao.GoodDao.GetGoodByID(goodID)
	if good.ID == 0 {
		return nil, response.CodeGoodsNotFound
	}
	// 3. 快速失败：库存（真实扣减在事务内，防并发超卖）
	if good.Stock <= 0 {
		return nil, response.CodeGoodsStockNotEnough
	}
	// 4. 快速失败：积分余额
	if user.Credit < good.Price {
		return nil, response.CodeCreditInsufficient
	}
	// 5. 事务：扣库存 → 扣积分 → 写订单（任一失败整体回滚）
	order := &model.Order{
		OrderNo:   generateOrderNo(),
		UserID:    userID,
		GoodsID:   good.ID,
		GoodsName: good.Name,
		Price:     good.Price,
		QQ:        *user.QQ,
		Nickname:  user.Nickname,
		CreatedAt: time.Now(),
	}
	err := global.LNF_DB.Transaction(func(tx *gorm.DB) error {
		// a. 条件更新扣库存（并发下仅 stock>0 生效；下架/删除后同样拒绝）
		affected, err := dao.GoodDao.RedeemStockTx(tx, good.ID)
		if err != nil {
			return err
		}
		if affected == 0 {
			return errGoodStockNotEnough
		}
		// b. 扣积分（AddUserCreditTx 内部行锁校验余额，不足时报错回滚；
		//    嵌套事务经 SAVEPOINT 实现，与 item 认领发分同款模式）
		if err := dao.UserDao.AddUserCreditTx(tx, userID, -good.Price,
			creditLogTypeShopRedeem, "积分兑换:"+good.Name, 0); err != nil {
			return err
		}
		// c. 写订单快照
		if err := dao.OrderDao.CreateOrderTx(tx, order); err != nil {
			return err
		}
		// 通知不进事务：NotificationService.Create 走全局连接，且通知失败不应回滚兑换；
		// 提交成功后统一发送（见下方 notifyRedeemNotifications）
		return nil
	})
	if err != nil {
		if errors.Is(err, errGoodStockNotEnough) {
			return nil, response.CodeGoodsStockNotEnough
		}
		if err.Error() == creditNotEnoughMsg {
			return nil, response.CodeCreditInsufficient
		}
		return nil, response.CodeDatabaseError
	}
	// 6. 事务提交成功后：重读余额（与响应 credit 同源）→ 同步发两条站内通知（失败仅记日志）
	//    → 异步发群消息（失败不影响兑换结果）
	after := dao.UserDao.GetUserByID(userID)
	notifyRedeemNotifications(order, after.Credit)
	go notifyRedeem(order)
	return &model.RedeemGoodsResponse{
		OrderNo:   order.OrderNo,
		GoodsID:   order.GoodsID,
		GoodsName: order.GoodsName,
		Price:     order.Price,
		Credit:    after.Credit,
		CreatedAt: order.CreatedAt,
	}, response.CodeSuccess
}

// notifyRedeemNotifications 兑换成功站内通知（事务提交后同步调用，先 A 后 B，失败仅记日志不影响兑换结果）：
// A. type=5 积分变动：含变动金额与事务后余额（与响应 credit 同源）；
// B. type=6 商品兑换：发货提醒，引导语与 chensong 群消息口径一致。
// 两条均 adminID=0 系统触发
func notifyRedeemNotifications(order *model.Order, balance int64) {
	if err := notificationService.Create(0, order.UserID, notificationTypeCreditChange,
		"积分变动提醒",
		fmt.Sprintf("你在积分商城兑换商品「%s」，积分 -%d，当前余额 %d 分。",
			order.GoodsName, order.Price, balance)); err != nil {
		log.Printf("[shop] 兑换积分变动通知发送失败 order_no=%s user_id=%d: %v",
			order.OrderNo, order.UserID, err)
	}
	if err := notificationService.Create(0, order.UserID, notificationTypeGoodsRedeem,
		"商品兑换成功",
		fmt.Sprintf("你已用 %d 积分兑换「%s」（订单号 %s），领取奖励请联系管理员～",
			order.Price, order.GoodsName, order.OrderNo)); err != nil {
		log.Printf("[shop] 商品兑换通知发送失败 order_no=%s user_id=%d: %v",
			order.OrderNo, order.UserID, err)
	}
}

// notifyRedeem 兑换成功后发送群消息到 activated_group（at 用户 QQ + 昵称 + 商品 + 引导联系管理员）
func notifyRedeem(order *model.Order) {
	msg := "[CQ:at,qq=" + order.QQ + "] " + order.Nickname +
		" 成功用 " + strconv.FormatInt(order.Price, 10) + " 积分兑换了「" + order.GoodsName +
		"」，领取奖励请联系管理员～"
	res, err := chensong.Client.SendGroupMessage(msg, global.LNF_CONFIG.ChenSong.ActivatedGroup)
	if err != nil || res.Status != "ok" {
		// 发送失败不影响兑换结果，仅忽略（与 QQGetCode 现有判定一致）
		return
	}
}
