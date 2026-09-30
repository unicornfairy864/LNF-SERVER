package advanced

import (
	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/handler"
	"github.com/unicornfairy864/LNF-SERVER/middleware"
)

type ShopRouter struct{}

func (s *ShopRouter) CreateRouter(api *gin.RouterGroup) {
	userGroup := api.Group("")
	// Public：商品浏览（无需登录）；list 静态路径与 :goodsID 参数路径并存（同 item 模块模式）
	public := userGroup.Group("/shop/goods")
	{
		public.GET("/list", handler.ShopHandler.ListGoodsHandler)
		public.GET("/:goodsID", handler.ShopHandler.GetGoodsHandler)
	}
	// Private：兑换与我的订单（:goodsID 参数名与 public 组保持一致，gin 要求同层级同名）
	private := userGroup.Group("/shop")
	private.Use(middleware.JWTAuthMiddleware())
	{
		private.POST("/goods/:goodsID/redeem", handler.ShopHandler.RedeemGoodsHandler)
		private.GET("/orders", handler.ShopHandler.ListMyOrdersHandler)
	}
	// Admin：商品管理（JWT + 管理员鉴权，role≥1）
	admin := userGroup.Group("/shop/goods")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.ServiceAdminAuthMiddleware())
	{
		admin.POST("/create", handler.ShopHandler.CreateGoodsHandler)
		admin.POST("/update", handler.ShopHandler.UpdateGoodsHandler)
		admin.POST("/delete", handler.ShopHandler.DeleteGoodsHandler)
	}
}
