package dao

import (
	"fmt"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserGroup struct{}

func (userGroup *UserGroup) GetUserByUsername(username string) (user *model.User) {
	user = &model.User{}
	global.LNF_DB.Where("username = ?", username).First(user)
	return user
}

func (userGroup *UserGroup) GetUserByID(id int64) (user *model.User) {
	user = &model.User{ID: id}
	global.LNF_DB.First(user)
	return user
}

func (userGroup *UserGroup) GetUserListByIDs(ids []int64) []model.User {
	var users []model.User
	global.LNF_DB.Where("id IN (?)", ids).Find(&users)
	return users
}

func (userGroup *UserGroup) GetUserByQQ(QQ string) (user *model.User) {
	user = &model.User{}
	global.LNF_DB.Where("qq = ?", QQ).First(user)
	return user
}

func (userGroup *UserGroup) CreateUser(user *model.User) (model.User, error) {
	err := global.LNF_DB.Create(user).Error
	return *user, err
}

func (userGroup *UserGroup) UpdateUserByVK(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	result := global.LNF_DB.Model(&model.User{}).
		Where(ConditionIDNotDeleted, id).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (userGroup *UserGroup) AddUserCredit(id int64, delta int64, logType int64, desc string, operatorId int64) error {
	return userGroup.AddUserCreditTx(global.LNF_DB, id, delta, logType, desc, operatorId)
}

// AddUserCreditTx 在指定 DB/事务上变更积分并写入流水（供外部事务复用，如认领关闭发分）
func (userGroup *UserGroup) AddUserCreditTx(db *gorm.DB, id int64, delta int64, logType int64, desc string, operatorId int64) error {
	if delta == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 加行锁读取当前用户积分，防止并发丢失更新。
		// 注意：Set("gorm:query_option", "FOR UPDATE") 是 GORM v1 写法，v2 已移除该语义且会静默忽略
		// （即修复前行锁从未生效）；v2 必须用 clause.Locking，且须在事务内执行才真正加锁。
		var user model.User
		if err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Where(ConditionIDNotDeleted, id).
			First(&user).Error; err != nil {
			return err
		}

		beforeAmount := user.Credit
		// 计算变动后的积分
		var afterAmount int64
		if delta <= -99999 {
			afterAmount = 0
		} else {
			afterAmount = beforeAmount + delta
			if afterAmount < 0 {
				return fmt.Errorf("credit not enough")
			}
		}
		realChange := afterAmount - beforeAmount

		// 更新用户积分
		updates := map[string]interface{}{
			"credit":     afterAmount,
			"updated_at": time.Now(),
		}
		if err := tx.Model(&model.User{}).
			Where(ConditionIDNotDeleted, id).
			Updates(updates).Error; err != nil {
			return err
		}

		// 写入积分变动记录
		log := &model.CreditLog{
			UserID:       id,
			ChangeAmount: realChange,
			BeforeAmount: beforeAmount,
			AfterAmount:  afterAmount,
			Type:         logType,
			Description:  desc,
			OperatorID:   operatorId,
			CreatedAt:    time.Now(),
		}
		if err := tx.Create(log).Error; err != nil {
			return err
		}
		return nil
	})
}

// creditLogListCols 积分流水列表显式列白名单：不含 operator_id / related_id / is_deleted。
// 既保证内部字段不外泄，也避免把 operator_id 的 NULL 扫描进非指针 int64 导致查询报错
const creditLogListCols = "id, user_id, change_amount, before_amount, after_amount, type, description, created_at"

// GetCreditLogsByUserID 分页查询用户积分流水（仅本人；is_deleted=0；q.Type 非 nil 时按类型筛选）。
// 排序 created_at DESC, id DESC：created_at 为秒级 DATETIME，同秒多行需 id 兜底保证翻页稳定；
// WHERE user_id + ORDER BY created_at 命中既有 idx_user_created，无需新增索引
func (userGroup *UserGroup) GetCreditLogsByUserID(userID int64, q *model.CreditLogListQuery) (logs []model.CreditLog, total int64, err error) {
	// 每次调用返回全新条件链，避免 Count 污染后续 Find
	base := func() *gorm.DB {
		db := global.LNF_DB.Model(&model.CreditLog{}).
			Where("user_id = ? AND is_deleted = 0", userID)
		if q.Type != nil {
			db = db.Where("type = ?", *q.Type)
		}
		return db
	}
	if err = base().Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = base().
		Select(creditLogListCols).
		Order("created_at DESC, id DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&logs).Error
	return logs, total, err
}
