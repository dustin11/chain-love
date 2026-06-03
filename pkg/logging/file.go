package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"senspace/pkg/setting"
	"time"
)

//var (
//	LogSavePath = "runtime/logs/"
//	LogSaveName = "log"
//	LogFileExt  = "log"
//	TimeFormat  = "20060102"
//)

func getLogFilePath() string {
	return fmt.Sprintf("%s", setting.Config.App.LogSavePath)
}

func getLogFileFullPath() string {
	prefixPath := getLogFilePath()
	suffixPath := fmt.Sprintf("%s%s.%s", setting.Config.App.LogSaveName, time.Now().Format(setting.Config.App.TimeFormat), setting.Config.App.LogFileExt)

	return filepath.Join(prefixPath, suffixPath)
}

func openLogFile(filePath string) *os.File {
	_, err := os.Stat(filePath)
	switch {
	case os.IsNotExist(err):
		mkDir()
	case os.IsPermission(err):
		if setting.IsDevLikeEnv() {
			log.Printf("logging fallback to stdout, permission denied for %s: %v", filePath, err)
			return nil
		}
		log.Fatalf("Permission :%v", err)
	}

	handle, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		if setting.IsDevLikeEnv() {
			log.Printf("logging fallback to stdout, failed to open %s: %v", filePath, err)
			return nil
		}
		log.Fatalf("Fail to OpenFile :%v", err)
	}

	return handle
}

func mkDir() {
	err := os.MkdirAll(filepath.Clean(getLogFilePath()), os.ModePerm)
	if err != nil {
		panic(err)
	}
}
