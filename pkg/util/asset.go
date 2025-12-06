package util

//
//import (
//	"spider/asset"
//	"fmt"
//	"io/ioutil"
//	"os"
//	"path/filepath"
//	"strings"
//)
//
//// AssetInfo loads and returns the asset info for the given name.
//// It returns an error if the asset could not be found or
//// could not be loaded.
//func AssetInfo(name string) (os.FileInfo, error) {
//	cannonicalName := strings.Replace(name, "\\", "/", -1)
//	if f, ok := asset._bindata[cannonicalName]; ok {
//		a, err := f()
//		if err != nil {
//			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
//		}
//		return a.info, nil
//	}
//	return nil, fmt.Errorf("AssetInfo %s not found", name)
//}
//// RestoreAsset restores an asset under the given directory
//func RestoreAsset(dir, name string) error {
//	data, err := asset.Asset(name)
//	if err != nil {
//		return err
//	}
//	info, err := asset.AssetInfo(name)
//	if err != nil {
//		return err
//	}
//	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
//	if err != nil {
//		return err
//	}
//	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
//	if err != nil {
//		return err
//	}
//	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
//	if err != nil {
//		return err
//	}
//	return nil
//}
//// RestoreAssets restores an asset under the given directory recursively
//func RestoreAssets(dir, name string) error {
//	children, err := asset.AssetDir(name)
//	// File
//	if err != nil {
//		return RestoreAsset(dir, name)
//	}
//	// Dir
//	for _, child := range children {
//		err = RestoreAssets(dir, filepath.Join(name, child))
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//func Restore(dirs [] string) bool {
//	isSuccess := true
//	for _, dir := range dirs {
//		// 解压dir目录到当前目录
//		if err := asset.RestoreAssets("./", dir); err != nil {
//			isSuccess = false
//			break
//		}
//	}
//	if !isSuccess {
//		for _, dir := range dirs {
//			os.RemoveAll(filepath.Join("./", dir))
//		}
//	}
//}
