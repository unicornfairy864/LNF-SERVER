package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/utils"
)

type SlHandler struct{}

func (sl *SlHandler) ReceiverHandler(c *gin.Context) {
	var body string
	if err := c.ShouldBindBodyWithPlain(&body); err != nil {
		response.Fail(c)
		return
	}
	utils.LogJson(body)
	response.Success(c)
}
