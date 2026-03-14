package dev_service

import (
	"encoding/json"
	"errors"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"chain-love/domain/dev"
	"chain-love/domain/dev/vo"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
)

func getPluginRoot(pluginId string) string {
	return filepath.Join(setting.Config.App.FilePath.Plugin, pluginId)
}

func GetPluginTree(pluginId string) (*vo.PluginFileNode, error) {
	rootPath := getPluginRoot(pluginId)
	info, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &vo.PluginFileNode{Name: pluginId, Type: "folder", Children: []*vo.PluginFileNode{}}, nil
		}
		return nil, err
	}

	return buildNode(rootPath, info), nil
}

func buildNode(path string, info fs.FileInfo) *vo.PluginFileNode {
	node := &vo.PluginFileNode{
		Name: info.Name(),
		Type: "file",
	}

	if info.IsDir() {
		node.Type = "folder"
		node.IsOpen = true
		node.Children = make([]*vo.PluginFileNode, 0)

		entries, err := os.ReadDir(path)
		if err == nil {
			for _, entry := range entries {
				childInfo, err := entry.Info()
				if err == nil {
					node.Children = append(node.Children, buildNode(filepath.Join(path, childInfo.Name()), childInfo))
				}
			}
		}
	} else {
		// Read file content
		content, err := os.ReadFile(path)
		if err == nil {
			node.Content = string(content)
		}
	}

	return node
}

func checkFileExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{
		".js":   true,
		".ts":   true,
		".json": true,
		".txt":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !allowed[ext] {
		return errors.New("unsupported file extension")
	}
	return nil
}

func UploadFile(pluginId string, relPath string, file *multipart.FileHeader) error {
	if err := checkFileExt(file.Filename); err != nil {
		return err
	}
	rootPath := getPluginRoot(pluginId)
    
	targetDir := filepath.Join(rootPath, filepath.Clean(relPath))
	if !strings.HasPrefix(targetDir, rootPath) {
		return errors.New("invalid path")
	}
	
	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return err
	}

	targetFile := filepath.Join(targetDir, filepath.Clean(file.Filename))
	if !strings.HasPrefix(targetFile, rootPath) {
		return errors.New("invalid path")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dest.Close()

	bytes := make([]byte, file.Size)
	_, err = src.Read(bytes)
	if err != nil {
		return err
	}
	_, err = dest.Write(bytes)
	return err
}

func AddFolder(pluginId string, relPath string) error {
	rootPath := getPluginRoot(pluginId)
	targetDir := filepath.Join(rootPath, filepath.Clean(relPath))
	if !strings.HasPrefix(targetDir, rootPath) {
		return errors.New("invalid path")
	}
	return os.MkdirAll(targetDir, 0700)
}

func Rename(pluginId string, oldPath string, newPath string) error {
	rootPath := getPluginRoot(pluginId)
	oldFullPath := filepath.Join(rootPath, filepath.Clean(oldPath))
	newFullPath := filepath.Join(rootPath, filepath.Clean(newPath))
	if !strings.HasPrefix(oldFullPath, rootPath) || !strings.HasPrefix(newFullPath, rootPath) {
		return errors.New("invalid path")
	}
	return os.Rename(oldFullPath, newFullPath)
}

func Delete(pluginId string, relPath string) error {
	rootPath := getPluginRoot(pluginId)
	fullPath := filepath.Join(rootPath, filepath.Clean(relPath))
	if !strings.HasPrefix(fullPath, rootPath) {
		return errors.New("invalid path")
	}
	return os.RemoveAll(fullPath)
}

func SavePlugin(ctx *contextx.AppContext, pluginId string, form *multipart.Form) (interface{}, error) {
	isNew := false
	var pid int64
	id, err := strconv.ParseInt(pluginId, 10, 64)
	if err != nil {
		isNew = true
	} else {
		pid = id
	}

	// 1. 寻找并解析 manifest.json 获取 version 和 name
	var version string
	var pluginName string
	manifestFound := false

	if files, ok := form.File["files"]; ok {
		for _, fileHeader := range files {
			if fileHeader.Filename == "manifest.json" {
				manifestFound = true
				src, err := fileHeader.Open()
				e.PanicIfErr(err)
				defer src.Close()

				bytes := make([]byte, fileHeader.Size)
				_, err = src.Read(bytes)
				e.PanicIfErr(err)

				var manifest struct {
					Version string `json:"version"`
					Name    string `json:"name"`
				}
				err = json.Unmarshal(bytes, &manifest)
				e.PanicIfErr(err)
				version = manifest.Version
				pluginName = manifest.Name
				break
			}
		}
	}

	if !manifestFound {
		e.PanicIfErr(errors.New("缺少manifest.json"))
	}
	if version == "" {
		version = "1.0.0"
	}
	if pluginName == "" {
		pluginName = "New Plugin"
	}

	var plugin dev.Plugin
	if isNew {
		plugin.Init(ctx.User)
		plugin.Name = pluginName
		plugin.Version = version
		err := plugin.Add()
		e.PanicIfErr(err)
		pid = plugin.Id
	} else {
		plugin.Id = pid
		plugin = plugin.GetById()
		if plugin.Id == 0 {
			e.PanicIfErr(errors.New("plugin not found"))
		}
		
		plugin.Name = pluginName
		plugin.Version = version
		err := plugin.Update(ctx.User.Id)
		e.PanicIfErr(err)
	}

	newPluginIdStr := strconv.FormatInt(pid, 10)
	
	// /plugin/{pluginId}/{version}
	rootPath := filepath.Join(getPluginRoot(newPluginIdStr), filepath.Clean(version))
	if !strings.HasPrefix(rootPath, getPluginRoot(newPluginIdStr)) {
		e.PanicIfErr(errors.New("invalid path"))
	}

	// Save all files from form
	if files, ok := form.File["files"]; ok {
		for _, fileHeader := range files {
			// check ext
			e.PanicIfErr(checkFileExt(fileHeader.Filename))

			relPath := filepath.Clean(fileHeader.Filename)
			fullPath := filepath.Join(rootPath, relPath)
			if !strings.HasPrefix(fullPath, rootPath) {
				e.PanicIfErr(errors.New("invalid path"))
			}

			dir := filepath.Dir(fullPath)
			err := os.MkdirAll(dir, 0700)
			e.PanicIfErr(err)

			src, err := fileHeader.Open()
			e.PanicIfErr(err)
			
			dest, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				src.Close()
				e.PanicIfErr(err)
			}
			bytes := make([]byte, fileHeader.Size)
			_, err = src.Read(bytes)
			if err != nil {
				src.Close()
				dest.Close()
				e.PanicIfErr(err)
			}
			_, err = dest.Write(bytes)
			if err != nil {
				src.Close()
				dest.Close()
				e.PanicIfErr(err)
			}
			src.Close()
			dest.Close()
		}
	}

	if isNew {
		return map[string]interface{}{"id": newPluginIdStr}, nil
	}
	return map[string]interface{}{"id": newPluginIdStr}, nil
}
