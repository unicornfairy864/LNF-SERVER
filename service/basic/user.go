package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/utils"
)

type UserServiceGroup struct{}

func (userService *UserServiceGroup) Create(c *gin.Context) {
	req := &model.CreateUserRequest{}
	c.ShouldBindBodyWithJSON(req)
	// 判断表单是否符合要求
	if (req.Username == "" || req.Nickname == "" || req.Password == "" ||
		len(req.Username) < 2 || len(req.Username) > 32 ||
		len(req.Nickname) < 2 || len(req.Nickname) > 32 ||
		len(req.Password) < 8 || len(req.Password) > 20) {
		// response.FailWithCode(c, response.CodeFormInvalid)
		response.FailWithData(c, response.CodeFormInvalid, req)
		return
	}
	if (dao.UserDao.GetUserByUsername(req.Username).ID != 0) {
		response.FailWithCode(c, response.CodeUsernameOccupied)
		return
	}
	user, err := dao.UserDao.CreateUser(&model.User{
		Username: req.Username,
		PasswordHash: utils.Bycrypt.GeneratePasswordHash(req.Password),
		Nickname: req.Nickname,
	})
	if err != nil {
		response.FailWithCode(c, response.CodeServerError)
		return
	}
	response.OKWithData(c, model.UserToResponse(&user))
}