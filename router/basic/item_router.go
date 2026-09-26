package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type ItemRouter struct{}

func (i *ItemRouter) CreateRouter(api *gin.RouterGroup) {
	userGroup := api.Group("")
	// Public
	public := userGroup.Group("/item")
	{
		public.GET("/list", handler.ItemHandler.ListItemHandler)
		public.GET("/:itemID", handler.ItemHandler.GetItemHandler)
	}
	// Private
	private := userGroup.Group("/item")
	private.Use(middleware.JWTAuthMiddleware())
	{
		private.POST("/create", handler.ItemHandler.CreateItemHandler)
		private.POST("/update", handler.ItemHandler.UpdateItemHandler)
		private.POST("/delete", handler.ItemHandler.DeleteItemHandler)
		private.GET("/mine", handler.ItemHandler.ListMyItemHandler)
		private.POST("/:itemID/images", handler.ItemHandler.SetItemImagesHandler)
		private.POST("/:itemID/claim", handler.ItemHandler.ClaimItemHandler)
		private.POST("/:itemID/claim/cancel", handler.ItemHandler.WithdrawClaimHandler)
		private.POST("/:itemID/confirm", handler.ItemHandler.ConfirmClaimHandler)
		private.POST("/:itemID/close", handler.ItemHandler.CloseSelfHandler)
	}
	// Admin（审核/状态流转等管理接口由后续 audit 模块补充）
	admin := userGroup.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
	}
}
