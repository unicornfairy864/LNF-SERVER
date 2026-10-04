package router

import (
	"github.com/unicornfairy864/LNF-SERVER/router/advanced"
	"github.com/unicornfairy864/LNF-SERVER/router/basic"
)

var (
	UserRouter         basic.UserRouter
	ItemRouter         basic.ItemRouter
	UploadRouter       basic.UploadRouter
	AnnouncementRouter basic.AnnouncementRouter

	LocationRouter     advanced.LocationRouter
	TagRouter          advanced.TagRouter
	NotificationRouter advanced.NotificationRouter
	AdminStatsRouter   advanced.AdminStatsRouter
	ShopRouter         advanced.ShopRouter
	AgentRouter        advanced.AgentRouter
	CommentRouter      advanced.CommentRouter
)
