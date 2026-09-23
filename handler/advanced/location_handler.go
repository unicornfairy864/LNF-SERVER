package advanced

import "github.com/gin-gonic/gin"

type LocationHandler struct{}

// CreateLocationHandler  管理员创建地点
// @Summary      管理员创建地点
// @Description  根据父元素ID创建地点，详细地址address可选，sort_order为递增排序条件
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateLocationRequest  true  "创建地点请求体"
// @Success      200      {object}  response.CommonResponse{}
// @Router       /api/v1/location/create [post]
func (l *LocationHandler) CreateLocationHandler(c *gin.Context) {
	return
}
