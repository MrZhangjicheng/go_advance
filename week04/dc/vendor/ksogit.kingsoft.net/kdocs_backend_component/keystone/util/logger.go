package util

import (
	"fmt"
	"os"

	"ksogit.kingsoft.net/kgo/log"
)

// 定义一个丢日志的回调函数：
func dropLogCallback(l *log.LogInstance, drop int) {
	// io线程chan满了，必需丢弃日志，请通过诸如普罗米修斯进行收集报警
	fmt.Printf("drop-log=%v, drop-sum=%v", l, drop)
}

// BuildLogger 构建logger
func BuildLogger(level string) {
	log.SetLevelS(level)
	// 给默认的IO线程设置一个丢日志时的回调通知。
	log.SetDropCallback(dropLogCallback)
	// 默认限制单条日志最大3k，可通过改此值
	log.MaxBytesPerLog = 1024 * 2

	logDir := os.Getenv("KAE_APP_LOG_DIR")
	hdlr, err := log.NewRotatingFileHandler(logDir+"/reform.log", log.MaxBytesPerLog, 5)
	if err != nil {
		log.Fatal("NewRotatingFileHandler: %v", err.Error())
	}
	log.AppendHandler(hdlr)
}
