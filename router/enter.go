package router

import (
	"github.com/unicornfairy864/LNF-SERVER/router/advanced"
	"github.com/unicornfairy864/LNF-SERVER/router/basic"
)

type Router struct {
	BasicRouter    basic.BasicRouterGroup
	AdvancedRouter advanced.AdvancedRouterGroup
}