package common

import (
	"log"
	"os"
)

var (
	InfoLog  = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile)
	WarnLog  = log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Lshortfile)
	ErrorLog = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
)

func Info(format string, v ...interface{})  { InfoLog.Printf(format, v...) }
func Warn(format string, v ...interface{})  { WarnLog.Printf(format, v...) }
func Error(format string, v ...interface{}) { ErrorLog.Printf(format, v...) }
