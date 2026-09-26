package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

// LocalStorage 本地磁盘存储：写入 {path}/年/月/日/{uuid}.{ext}，返回 /uploads/... 相对 URL
type LocalStorage struct{}

// Save 落盘并返回相对 URL
func (s *LocalStorage) Save(data []byte, ext string) (string, error) {
	cfg := global.LNF_CONFIG.Storage
	now := time.Now()

	// 磁盘路径用 filepath.Join 适配操作系统，URL 部分显式用 / 拼接
	relDir := fmt.Sprintf("%04d/%02d/%02d", now.Year(), int(now.Month()), now.Day())
	absDir := filepath.Join(cfg.Path, relDir)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", fmt.Errorf("create upload dir %s: %w", absDir, err)
	}

	name := uuid.NewString() + "." + ext
	if err := os.WriteFile(filepath.Join(absDir, name), data, 0o644); err != nil {
		return "", fmt.Errorf("write upload file %s: %w", name, err)
	}
	return URLPrefix + "/" + relDir + "/" + name, nil
}
