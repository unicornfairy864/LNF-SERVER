package basic

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unicornfairy864/LNF-SERVER/global"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service"
)

type UploadHandlerGroup struct{}

// uploadMultipartOverhead multipart 报文较文件本体的额外开销余量（字节）
const uploadMultipartOverhead = 1024

// UploadImageHandler 登录用户上传图片
// @Summary      登录用户上传图片
// @Description  头像与物品图通用的图片上传接口，multipart/form-data 的 file 字段；<br />仅支持 jpg/png/webp，单文件最大 5MB（storage.max_size 可调）；返回相对 URL（如 /uploads/2026/09/26/xxx.jpg），入库与前端展示时需自行拼接服务器源；<br />文件过大返回 9，类型不支持返回 10，保存失败返回 11
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "图片文件"
// @Success      200  {object}  response.CommonResponse{data=model.UploadImageResponse}
// @Router       /api/v1/upload/image [post]
func (uploadHandler *UploadHandlerGroup) UploadImageHandler(c *gin.Context) {
	cfg := global.LNF_CONFIG.Storage

	// 1. Content-Length 预检：超限直接拒，不读包体
	if c.Request.ContentLength > cfg.MaxSize+uploadMultipartOverhead {
		response.FailWithCode(c, response.CodeUploadFileTooLarge)
		return
	}
	// 2. 硬上限兜底：防伪造 Content-Length 或 chunked 传输绕过预检
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxSize+uploadMultipartOverhead)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		// 触发 MaxBytesReader 上限后，multipart 解析错误固定含 request body too large
		if strings.Contains(err.Error(), "request body too large") {
			response.FailWithCode(c, response.CodeUploadFileTooLarge)
			return
		}
		response.FailWithCode(c, response.CodeParamError)
		return
	}

	res, code := service.UploadService.UploadImageService(fileHeader)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, res)
}
