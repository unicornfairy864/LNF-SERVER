package basic

import (
	"github.com/gin-gonic/gin"
)

type UserServiceGroup struct{}

func (userService *UserServiceGroup) Create(c *gin.Context) {
	c.JSON(200, "Message: Test")
}