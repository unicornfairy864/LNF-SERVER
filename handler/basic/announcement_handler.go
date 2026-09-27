package basic

import (
	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type AnnouncementHandlerGroup struct{}

func (a *AnnouncementHandlerGroup) CreateHandler(c *gin.Context) {
	rep := model.AnnouncementUpdateRequest{}
	if err := c.ShouldBindJSON(&rep); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	err := service.AnnouncementService.CreateAnnouncement(&rep)
	if err != response.CodeSuccess {
		response.FailWithCode(c, err)
	}
	response.Success(c)
}

func (a *AnnouncementHandlerGroup) UpdatedHandler(c *gin.Context) {
	rep := model.AnnouncementUpdateRequest{}
	if err := c.ShouldBindJSON(&rep); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	err := service.AnnouncementService.UpdatedAnnouncement(&rep)
	if err != response.CodeSuccess {
		response.FailWithCode(c, err)
	}
	response.Success(c)
}

func (a *AnnouncementHandlerGroup) GetHandler(c *gin.Context) {
	req := model.AnnouncementGetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	req.Auth = false
	response.SuccessWithData(c, service.AnnouncementService.GetAnnouncements(&req))
}

func (a *AnnouncementHandlerGroup) AuthGetHandler(c *gin.Context) {
	req := model.AnnouncementGetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeParamError)
	}
	response.SuccessWithData(c, service.AnnouncementService.GetAnnouncements(&req))
}
