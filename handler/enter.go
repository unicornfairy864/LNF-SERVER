package handler

import (
	"github.com/unicornfairy864/LNF-SERVER/handler/advanced"
	"github.com/unicornfairy864/LNF-SERVER/handler/basic"
)

var (
	UserHandler basic.UserHandlerGroup
	ItemHandler basic.ItemHandlerGroup

	LocationHandler advanced.LocationHandler
)
