package service

import (
	"github.com/unicornfairy864/LNF-SERVER/service/advanced"
	"github.com/unicornfairy864/LNF-SERVER/service/basic"
)

var (
	UserService     basic.UserServiceGroup
	ItemService     basic.ItemServiceGroup
	TagService      advanced.TagServiceGroup
	LocationService advanced.LocationServiceGroup
)
