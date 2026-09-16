package response

import "github.com/gin-gonic/gin"

type CommonResponse struct {
	Code    Code        `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Result(c *gin.Context, code Code, message string, data interface{}) {
	c.JSON(200, CommonResponse{
		Code: code,
		Message: message,
		Data: data,
	})
}

func Success(c *gin.Context) {
	Result(c, CodeSuccess, Msg[CodeSuccess], map[string]interface{}{})
}

func SuccessWithData(c *gin.Context, data interface{}) {
	Result(c, CodeSuccess, Msg[CodeSuccess], data)
}

func Fail(c *gin.Context) {
	Result(c, CodeUnknownError, Msg[CodeUnknownError], map[string]interface{}{})
}

func FailWithCode(c *gin.Context, code Code) {
	Result(c, code, Msg[code], map[string]interface{}{})
}

func FailWithData(c *gin.Context, code Code, data interface{}) {
	Result(c, code, Msg[code], data)
} 

func TestWithData(c *gin.Context, data interface{}) {
	Result(c, CodeTest, Msg[CodeTest], data)
}