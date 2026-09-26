package basic

import (
	"io"
	"mime/multipart"
	"net/http"
	"slices"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
	"github.com/unicornfairy864/LNF-SERVER/service/storage"
)

type UploadServiceGroup struct{}

// mimeToExt 支持的图片类型：嗅探出的 MIME -> 落盘扩展名
// 与 config.yaml 的 storage.allowed_types 联合生效：新增类型需两处同步
var mimeToExt = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

// UploadImageService 校验并保存上传图片（头像 / 物品图通用），返回相对 URL
func (uploadService *UploadServiceGroup) UploadImageService(fileHeader *multipart.FileHeader) (*model.UploadImageResponse, response.Code) {
	cfg := global.LNF_CONFIG.Storage

	// 1. 大小预检：multipart 头部声明的尺寸可伪造，仅作快速失败
	if fileHeader.Size > cfg.MaxSize {
		return nil, response.CodeUploadFileTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, response.CodeUploadFailed
	}
	defer file.Close()

	// 2. 限读 max_size+1 字节读入内存：读满即超限，防止伪造大小绕过预检
	data, err := io.ReadAll(io.LimitReader(file, cfg.MaxSize+1))
	if err != nil {
		return nil, response.CodeUploadFailed
	}
	if int64(len(data)) > cfg.MaxSize {
		return nil, response.CodeUploadFileTooLarge
	}

	// 3. 嗅探文件头判断真实类型：不信任扩展名与客户端声明的 Content-Type
	mimeType := http.DetectContentType(data)
	ext, ok := mimeToExt[mimeType]
	if !ok || !slices.Contains(cfg.AllowedTypes, mimeType) {
		return nil, response.CodeUploadFileTypeInvalid
	}

	// 4. 落盘并返回相对 URL
	url, err := storage.Uploader.Save(data, ext)
	if err != nil {
		return nil, response.CodeUploadFailed
	}
	return &model.UploadImageResponse{URL: url}, response.CodeSuccess
}
