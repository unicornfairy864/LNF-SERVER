package config

// StorageConfig 上传文件存储配置
type StorageConfig struct {
	Driver       string   `mapstructure:"driver"`        // 存储后端，当前仅 local
	Path         string   `mapstructure:"path"`          // 本地存储根目录（相对服务运行目录）
	MaxSize      int64    `mapstructure:"max_size"`      // 单文件大小上限（字节）
	AllowedTypes []string `mapstructure:"allowed_types"` // 允许的 MIME 白名单，需与 service/basic/upload_service.go 的 mimeToExt 保持同步
}
