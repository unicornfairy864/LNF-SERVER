package dao

import (
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

type UserGroup struct{}

func (userGroup *UserGroup) GetUserByUsername(username string) (user *model.User) {
	user = &model.User{Username: username}
	global.LNF_DB.Where("username = ?", username).First(user)
	return user
}

func (userGroup *UserGroup) CreateUser (user *model.User) (model.User, error) {
	err := global.LNF_DB.Create(user).Error
	return *user, err
}