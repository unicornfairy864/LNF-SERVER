package storage

// URLPrefix 上传文件的 URL 前缀；路由静态目录注册与 URL 生成共用此常量，保证一致
const URLPrefix = "/uploads"

// Storage 存储后端接口；未来切换对象存储（COS/OSS）时新增实现即可，业务代码零改动
type Storage interface {
	// Save 保存图片二进制内容，ext 为不含点号的扩展名（如 jpg），返回可直接入库的相对 URL
	Save(data []byte, ext string) (string, error)
}

// New 构造存储后端
// LocalStorage 无状态且在调用期读取配置，因此在 Viper 加载前构造也安全
func New() Storage {
	// 当前仅实现 local；未来接入 COS/OSS 时在此按 storage.driver 分发
	return &LocalStorage{}
}

// Uploader 全局存储实例
var Uploader = New()
