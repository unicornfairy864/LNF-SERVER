package dao

import (
	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/model/basic"
)

type UserGroup struct{}

func (userGroup *UserGroup) UserExistsById(id int64) bool {
	var count int64
	global.LNF_DB.Model(&model.User{}).Where("id = ?", id).Count(&count)
	return count > 0
}

func (userGroup *UserGroup) GetUserById(id int64) (user *model.User) {
	user = &model.User{ID: id}
	global.LNF_DB.First(user)
	return user
}

func (userGroup *UserGroup) CreateUser (user *model.User) (model.User, error) {
	err := global.LNF_DB.Create(user).Error
	return *user, err
}