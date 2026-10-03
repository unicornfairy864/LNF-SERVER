package service

import (
	"github.com/unicornfairy864/LNF-SERVER/service/advanced"
	"github.com/unicornfairy864/LNF-SERVER/service/basic"
)

var (
	UserService     basic.UserServiceGroup
	ItemService     basic.ItemServiceGroup
	UploadService   basic.UploadServiceGroup
	TagService      advanced.TagServiceGroup
	LocationService advanced.LocationServiceGroup
	ShopService     advanced.ShopServiceGroup

	AnnouncementService basic.AnnouncementServiceGroup
	NotificationService advanced.NotificationServiceGroup
	AdminStatsService   advanced.AdminStatsServiceGroup
)
