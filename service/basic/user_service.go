package basic

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/utils"
)

type UserServiceGroup struct{}

func (userService *UserServiceGroup) Create(req *model.CreateUserRequest) (*model.User, response.Code) {
	// 判断表单是否符合要求
	if (req.Username == "" || req.Nickname == "" || req.Password == "" ||
		len(req.Username) < 2 || len(req.Username) > 32 ||
		len(req.Nickname) < 2 || len(req.Nickname) > 32 ||
		len(req.Password) < 8 || len(req.Password) > 20) {
		return nil, response.CodeParamError
	}
	// 用户名是否被占用
	if (dao.UserDao.GetUserByUsername(req.Username).ID != 0) {
		return nil, response.CodeUsernameOccupied
	}
	user, err := dao.UserDao.CreateUser(&model.User{
		Username: req.Username,
		PasswordHash: utils.Bycrypt.GeneratePasswordHash(req.Password),
		Nickname: req.Nickname,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return nil, response.CodeServerError
	}
	return &user, response.CodeSuccess
}

func (userService *UserServiceGroup) Login(req *model.LoginRequest) (*model.User, *string, response.Code) {
	// 判断表单是否符合要求
	if (req.Username == "" || req.Password == "" ||
		len(req.Username) < 2 || len(req.Username) > 32 ||
		len(req.Password) < 8 || len(req.Password) > 20) {
		return nil, nil, response.CodeFormInvalid
	}
	// 用户是否存在
	user := dao.UserDao.GetUserByUsername(req.Username)
	if (user.ID == 0) {
		return nil, nil, response.CodeUserOrPasswordError
	}
	if (!utils.Bycrypt.CheckPassword(req.Password, user.PasswordHash)) {
		return nil, nil, response.CodeUserOrPasswordError
	}
	token, err := utils.JWT.GenerateToken(user)
	if (err != nil) {
		return nil, nil, response.CodeServerError
	}
	return user, &token, response.CodeSuccess
}