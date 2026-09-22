package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type SlHandler struct{}

func (sl *SlHandler) ReceiverHandler(c *gin.Context) {
	var body string
	if err := c.ShouldBindBodyWithPlain(&body); err != nil {
		response.Fail(c)
		return
	}
	response.Success(c)
}
