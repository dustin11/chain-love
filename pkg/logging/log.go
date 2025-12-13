package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type Level int

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 1

	logger     *log.Logger
	logPrefix  = ""
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	//日志中需要跳过的通用异常调用 key:代码文件 value:跳过步骤
	//从最外层开始检查，没有调用的再检查后面的内层，找到匹配的调用，跳过步骤+1即为异常调用处
	skipProcess = map[string]int{
		PROC_ERROR:         4,
		PROC_ERROR_HANDLER: 3,
	}
)

func getSkipStep(level Level, key string) int {
	if level != ERROR {
		return DefaultCallerDepth
	}
	if key == "" {
		return skipProcess[PROC_ERROR_HANDLER]
	}
	if v, ok := skipProcess[key]; ok {
		return v
	}
	return skipProcess[PROC_ERROR_HANDLER]
}

/*
执行路径，从里到外
从err_handler.go抛出路径
2020/02/10 16:38:51 F:/project/go/gin-hello/pkg/logging/log.go
2020/02/10 16:38:51 F:/project/go/gin-hello/pkg/logging/log.go
2020/02/10 16:38:51 F:/project/go/gin-hello/middleware/err_handler.go
2020/02/10 16:38:51 D:/Program Files/go/src/runtime/panic.go
2020/02/10 16:38:51 F:/project/go/gin-hello/middleware/auth.go
2020/02/10 16:38:51 F:/project/go/gin-hello/middleware/auth.go
未处理的异常路径
2020/02/10 16:59:10 F:/project/go/gin-hello/pkg/logging/log.go
2020/02/10 16:59:10 F:/project/go/gin-hello/pkg/logging/log.go
2020/02/10 16:59:10 F:/project/go/gin-hello/middleware/err_handler.go
2020/02/10 16:59:10 D:/Program Files/go/src/runtime/panic.go
2020/02/10 16:59:10 D:/Program Files/go/src/runtime/panic.go
2020/02/10 16:59:10 F:/project/go/gin-hello/handler/api/api_v1/user.go
从error.go抛出路径
2020/02/10 16:44:15 F:/project/go/gin-hello/pkg/logging/log.go
2020/02/10 16:44:15 F:/project/go/gin-hello/pkg/logging/log.go
2020/02/10 16:44:15 F:/project/go/gin-hello/middleware/err_handler.go
2020/02/10 16:44:15 D:/Program Files/go/src/runtime/panic.go
2020/02/10 16:44:15 F:/project/go/gin-hello/pkg/e/error.go
2020/02/10 16:44:15 F:/project/go/gin-hello/service/user_service/user.go
2020/02/10 16:44:15 F:/project/go/gin-hello/handler/api/api_v1/home.go

非error级别
2020/02/11 11:05:25 F:/project/go/gin-hello/pkg/logging/log.go
2020/02/11 11:05:25 F:/project/go/gin-hello/handler/api/api_v1/home.go
2020/02/11 11:05:25 D:/Program Files/go/self/pkg/mod/github.com/gin-gonic/gin@v1.4.0/context.go
*/

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

func Setup() {
	filePath := getLogFileFullPath()
	F = openLogFile(filePath)

	logger = log.New(F, DefaultPrefix, log.LstdFlags)
}

func print(v ...interface{}) {
	log.Println(v)
	logger.Println(v)
}

func Debug(v ...interface{}) {
	setPrefix(DEBUG, "")
	log.Println(v)
	logger.Println(v)
}

func Info(v ...interface{}) {
	setPrefix(INFO, "")
	log.Println(v)
	logger.Println(v)
}

func Warn(v ...interface{}) {
	setPrefix(WARNING, "")
	log.Println(v)
	logger.Println(v)
}

func Error(v ...interface{}) {
	setPrefix(ERROR, PROC_ERROR)
	log.Println(v)
	logger.Println(v)
}

func ErrorLocate(statCode int, msg string, locate string) {
	setPrefix(ERROR, locate)
	log.Println(msg)
	logger.Println(statCode, msg)
}

func Fatal(v ...interface{}) {
	setPrefix(FATAL, "")
	log.Println(v)
	logger.Fatalln(v)
}

func setPrefix(level Level, locate string) {
	//for j := 0; j < 20; j++ {
	//	_, file, _, _ := runtime.Caller(j)
	//	log.Println(file)
	//}
	step := getSkipStep(level, locate) + 1
	//_, file, line, ok := getRealErrorStackCaller(locate)
	if _, file, line, ok := runtime.Caller(step); ok {
		if filepath.Base(file) == "panic.go" {
			//未处理，未手动panic的异常
			_, file, line, ok = runtime.Caller(step + 1)
		}
		logPrefix = fmt.Sprintf("[%s][%s:%d]", levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s]", levelFlags[level])
	}
	logger.SetPrefix(logPrefix)
}

////获取真正异常调用点（跳过通用异常调用）
//func getRealErrorStackCaller(locate string) (pc uintptr, file string, line int, ok bool) {
//
//	if pc, file, line, ok = runtime.Caller(DefaultCallerDepth); !ok{
//		return 0, "", 0, false
//	}
//	step := getSkipStep(locate) +1
//	if pc, file, line, ok = runtime.Caller(step); !ok{
//		return 0, "", 0, false
//	}
//
//	//取出map的key
//	skipKey := make([]string, len(skipProcess))
//	k :=0
//	for key := range skipProcess{
//		skipKey[k] = key
//		k++
//	}
//	for j:=0;j<20;j++{
//		pc, file, line, ok = runtime.Caller(j)
//		log.Println(file)
//	}
//	log.Println("... ... ... ... ... ...")
//	level := 0
//	path := file
//	for i:= 0;i<len(skipKey);i++ {
//		//file值已变，说明上个循环找到匹配，换path再从数组开头找
//		if path != file {
//			i =0
//			path = file
//		}
//		if strings.HasSuffix(path, skipKey[i]){
//			//加上map中跳过函数栈数
//			level = level + skipProcess[skipKey[i]]
//			//如使用递归，调用层次会一直增加
//			if pc, file, line, ok = runtime.Caller(level); !ok{
//				return 0, "", 0, false
//			}
//		}
//	}
//	//找完了没找到，返回已找到结果
//	return pc, file, line, ok
//}
