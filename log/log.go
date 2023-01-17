package log

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	logFile, err := os.OpenFile("./replay.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Panic("打开日志文件异常")
	}
	Logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
}
