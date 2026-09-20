package chensong

import (
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/client"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/handler"
	"github.com/unicornfairy864/LNF-SERVER/chensong/internal/middleware"
)

// ÕâÊÇ³ÂËÉ»úÆ÷ÈËÔÚ±¾ÏîÄ¿µÄÖ÷ÒªÊµÏÖ

var (
	Client       client.ChenSongClient
	SlHandler    handler.SlHandler
	SlMiddleware middleware.SlMiddleware
)
