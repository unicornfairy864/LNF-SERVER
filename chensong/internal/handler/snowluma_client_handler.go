package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type SlHandler struct{}

func (sl *SlHandler) ReceiverHandler(c *gin.Context) {
	var str string
	err := c.ShouldBindBodyWithPlain(&str)
	if err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	// Receive Service
}
