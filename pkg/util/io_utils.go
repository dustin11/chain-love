package util

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func RootPath() string {
	s, err := exec.LookPath(os.Args[0])
	if err != nil {
		log.Panicln("发生错误", err.Error())
	}
	i := strings.LastIndex(s, "\\")
	path := s[0 : i+1]
	return path
}

func CreateDirIfNotExits(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		//必须分成两步：先创建文件夹、再修改权限
		os.MkdirAll(path, os.ModePerm) //0777也可以os.ModePerm
		os.Chmod(path, 0777)
		return true
	}
	return false
}

// RemoveIfExists 删除文件，如果文件不存在返回 nil，其他错误向上返回
func RemoveIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Remove(path)
}
