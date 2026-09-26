package router

import (
	"github.com/unicornfairy864/LNF-SERVER/router/advanced"
	"github.com/unicornfairy864/LNF-SERVER/router/basic"
)

var (
	UserRouter   basic.UserRouter
	ItemRouter   basic.ItemRouter
	UploadRouter basic.UploadRouter

	LocationRouter advanced.LocationRouter
	TagRouter      advanced.TagRouter
)
