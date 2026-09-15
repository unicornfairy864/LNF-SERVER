package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type UserServiceGroup struct{}

func (userService *UserServiceGroup) Create(c *gin.Context) {
	response.OK(c)
}