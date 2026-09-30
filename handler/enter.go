package handler

import (
	"github.com/unicornfairy864/LNF-SERVER/handler/advanced"
	"github.com/unicornfairy864/LNF-SERVER/handler/basic"
)

var (
	UserHandler         basic.UserHandlerGroup
	ItemHandler         basic.ItemHandlerGroup
	UploadHandler       basic.UploadHandlerGroup
	AnnouncementHandler basic.AnnouncementHandlerGroup

	LocationHandler     advanced.LocationHandler
	TagHandler          advanced.TagHandler
	NotificationHandler advanced.NotificationHandlerGroup
	ShopHandler         advanced.ShopHandlerGroup
)
