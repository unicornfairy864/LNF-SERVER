package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type CommentHandlerGroup struct{}

func (c *CommentHandlerGroup) CreateHandler(g *gin.Context) {
	req := model.CreateCommentRequest{}
	if err := g.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	req.UserID = g.GetInt64(middleware.ContextID)
	err := service.CommentService.CreateComment(&req)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.Success(g)
}

func (c *CommentHandlerGroup) UpdateHandler(g *gin.Context) {
	req := model.UpdateCommentRequest{}
	if err := g.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	err := service.CommentService.UpdateComment(&req)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.Success(g)
}

func (c *CommentHandlerGroup) GetListHandler(g *gin.Context) {
	qreq := model.CommentGetlistRequestQuery{}
	preq := model.ItemRequestParam{}
	if err := g.ShouldBindQuery(&qreq); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	if err := g.ShouldBindUri(&preq); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	list, err := service.CommentService.GetList(&qreq, &preq)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.SuccessWithData(g, list)
}

func (c *CommentHandlerGroup) GetChildHandler(g *gin.Context) {
	qreq := model.GetChildrenRequestQuery{}
	preq := model.ItemRequestParam{}
	if err := g.ShouldBindQuery(&qreq); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	if err := g.ShouldBindUri(&preq); err != nil {
		response.FailWithCode(g, response.CodeParamError)
		return
	}
	list, err := service.CommentService.GetChildComments(&qreq, &preq)
	if err != response.CodeSuccess {
		response.FailWithCode(g, err)
		return
	}
	response.SuccessWithData(g, list)
}
