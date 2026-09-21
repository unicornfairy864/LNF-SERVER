package dao

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
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
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
