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
