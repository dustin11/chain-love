package dev_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func compareVersions(v1, v2 string) int {
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")
	for i := 0; i < len(p1) && i < len(p2); i++ {
		n1, _ := strconv.Atoi(p1[i])
		n2, _ := strconv.Atoi(p2[i])
		if n1 != n2 {
			return n1 - n2
		}
	}
	return len(p1) - len(p2)
}

func GetLatestVersion(pluginId string) string {
	rootPath := getPluginRoot(pluginId)
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return "1.0.0"
	}
	var maxVer string
	for _, entry := range entries {
		if entry.IsDir() {
			v := entry.Name()
			if maxVer == "" || compareVersions(v, maxVer) > 0 {
				maxVer = v
			}
		}
	}
	if maxVer == "" {
		return "1.0.0"
	}
	return maxVer
}

func getLatestPluginRoot(pluginId string) string {
	return filepath.Join(getPluginRoot(pluginId), GetLatestVersion(pluginId))
}

func checkFileExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	e.PanicIfParameterError(ext == "", "不支持此文件类型！")
	// if ext == "" {
	// 	return errors.New("unsupported file extension: [none]")
	// }
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
	e.PanicIfParameterError(!allowed[ext], fmt.Sprintf("不支持文件类型：%s", ext))
	// if !allowed[ext] {
	// 	return errors.New("unsupported file extension: " + ext)
	// }
	return nil
}

func cleanAndCheckPath(baseDir, userInputPath string) (string, error) {
	clean := filepath.Clean(userInputPath)
	fullPath := filepath.Join(baseDir, clean)
	if !strings.HasPrefix(fullPath, baseDir) {
		return "", errors.New("invalid path: out of boundary")
	}
	return fullPath, nil
}

func GetPluginTree(pluginId string) (*vo.PluginFileNode, error) {
	rootPath := getLatestPluginRoot(pluginId)

	var pluginName string
	if id, err := strconv.ParseInt(pluginId, 10, 64); err == nil {
		plugin := dev.Plugin{Id: id}.GetById()
		pluginName = plugin.Name
	}
	if pluginName == "" {
		pluginName = pluginId
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(filepath.Join(rootPath, "assets"), 0700)
			return &vo.PluginFileNode{Name: pluginName, Type: "folder", Children: []*vo.PluginFileNode{}}, nil
		}
		return nil, err
	}

	node := buildNode(rootPath, info)
	node.Name = pluginName
	return node, nil
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
		content, err := os.ReadFile(path)
		if err == nil {
			node.Content = string(content)
		}
	}

	return node
}

func UploadFile(pluginId string, relPath string, file *multipart.FileHeader) error {
	if err := checkFileExt(file.Filename); err != nil {
		return err
	}
	rootPath := getLatestPluginRoot(pluginId)

	targetDir, err := cleanAndCheckPath(rootPath, relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return err
	}

	targetFile, err := cleanAndCheckPath(targetDir, file.Filename)
	if err != nil {
		return err
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

	_, err = io.Copy(dest, src)
	return err
}

func AddFolder(pluginId string, relPath string) error {
	rootPath := getLatestPluginRoot(pluginId)
	targetDir, err := cleanAndCheckPath(rootPath, relPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(targetDir, 0700)
}

func Rename(pluginId string, oldPath string, newPath string) error {
	rootPath := getLatestPluginRoot(pluginId)
	oldFullPath, err := cleanAndCheckPath(rootPath, oldPath)
	if err != nil {
		return err
	}
	newFullPath, err := cleanAndCheckPath(rootPath, newPath)
	if err != nil {
		return err
	}
	return os.Rename(oldFullPath, newFullPath)
}

func Delete(pluginId string, relPath string) error {
	rootPath := getLatestPluginRoot(pluginId)
	fullPath, err := cleanAndCheckPath(rootPath, relPath)
	if err != nil {
		return err
	}
	return os.RemoveAll(fullPath)
}

func DeletePlugin(pluginId string) error {
	if id, err := strconv.ParseInt(pluginId, 10, 64); err == nil {
		dev.Plugin{Id: id}.Delete()
	}
	rootPath := getPluginRoot(pluginId)
	return os.RemoveAll(rootPath)
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

	var version string
	var pluginName string
	manifestFound := false

	if files, ok := form.File["files"]; ok {
		for _, fileHeader := range files {
			if filepath.Base(filepath.Clean(fileHeader.Filename)) == "manifest.json" {
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

	rootPath, err := cleanAndCheckPath(getPluginRoot(newPluginIdStr), version)
	e.PanicIfErr(err)

	err = os.MkdirAll(filepath.Join(rootPath, "assets"), 0700)
	e.PanicIfErr(err)

	if folders, ok := form.Value["folders"]; ok {
		for _, folder := range folders {
			dir, err := cleanAndCheckPath(rootPath, folder)
			if err == nil {
				os.MkdirAll(dir, 0700)
			}
		}
	}

	if files, ok := form.File["files"]; ok {
		for _, fileHeader := range files {
			e.PanicIfErr(checkFileExt(fileHeader.Filename))

			fullPath, err := cleanAndCheckPath(rootPath, fileHeader.Filename)
			e.PanicIfErr(err)

			dir := filepath.Dir(fullPath)
			err = os.MkdirAll(dir, 0700)
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

	return map[string]interface{}{"id": newPluginIdStr}, nil
}
