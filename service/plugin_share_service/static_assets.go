package plugin_share_service

import (
	"errors"
	"fmt"
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
	target, err := resolveBackgroundPath(root, key)
	if err != nil {
		return err
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
	// CreateTemp 默认生成 0600 文件，原子发布前转为 Nginx 工作进程可读。
	if err := os.Chmod(temporaryPath, 0o644); err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}

func removeBackground(key string) error {
	if cleanKey := strings.TrimSpace(key); cleanKey == "" || cleanKey == "." {
		return nil
	}
	target, err := resolveBackgroundPath(SharedStaticRoot(), key)
	if err != nil {
		return err
	}
	err = os.Remove(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// ensureBackgroundReadable 在返回历史分享时修复 CreateTemp 遗留的 0600 权限。
func ensureBackgroundReadable(key string) error {
	target, err := resolveBackgroundPath(SharedStaticRoot(), key)
	if err != nil {
		return err
	}
	info, err := os.Lstat(target)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("分享背景不是普通文件: %s", key)
	}
	if info.Mode().Perm() == 0o644 {
		return nil
	}
	return os.Chmod(target, 0o644)
}

func resolveBackgroundPath(root string, key string) (string, error) {
	cleanRoot := filepath.Clean(root)
	cleanKey := strings.TrimSpace(key)
	if cleanKey == "" || cleanKey != key || cleanKey == "." || cleanKey == ".." || filepath.IsAbs(cleanKey) || filepath.Clean(cleanKey) != cleanKey || filepath.Base(cleanKey) != cleanKey || strings.ContainsAny(cleanKey, `/\\`) {
		return "", errors.New("分享背景路径无效")
	}
	target := filepath.Join(cleanRoot, cleanKey)
	relative, err := filepath.Rel(cleanRoot, target)
	if err != nil || relative != cleanKey {
		return "", errors.New("分享背景路径无效")
	}
	return target, nil
}
