package plugin_share_service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"senspace/pkg/setting"
)

// SharedStaticRoot 返回分享背景的运行时静态目录。
func SharedStaticRoot() string {
	if base := strings.TrimSpace(setting.Config.App.RuntimeRootPath); base != "" {
		return filepath.Join(base, "plugin-shared")
	}
	return filepath.Join("runtime", "plugin-shared")
}

func writeBackgroundAtomic(key string, data []byte) error {
	root := SharedStaticRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	target := filepath.Join(root, key)
	if filepath.Dir(target) != filepath.Clean(root) {
		return errors.New("分享背景路径无效")
	}
	temporary, err := os.CreateTemp(root, ".plugin-share-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}

func removeBackground(key string) error {
	key = filepath.Base(strings.TrimSpace(key))
	if key == "" || key == "." {
		return nil
	}
	err := os.Remove(filepath.Join(SharedStaticRoot(), key))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
