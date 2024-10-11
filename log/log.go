package log

import (
	syslog "log"
	"os"
)

var (
	log    = New()
	red    = "\033[31m"
	yellow = "\033[33m"
	rest   = "\033[0m"
)

type Logger struct {
	infoLogger  *syslog.Logger
	warnLogger  *syslog.Logger
	errorLogger *syslog.Logger
}

func Info(format string, v ...any) {
	log.infoLogger.Printf(format+"\n", v...)
}

func Warn(format string, v ...any) {
	log.warnLogger.Printf(format+rest+"\n", v...)
}

func Error(format string, v ...any) {
	log.errorLogger.Printf(format+rest+"\n", v...)
}

func Panic(format string, v ...any) {
	log.errorLogger.Printf(format+rest+"\n", v...)
	os.Exit(1)
}

func New() (log *Logger) {
	log = new(Logger)
	log.infoLogger = syslog.New(os.Stdout, "", syslog.LstdFlags)
	log.errorLogger = syslog.New(os.Stderr, red+"", syslog.LstdFlags)
	log.warnLogger = syslog.New(os.Stdout, yellow+"", syslog.LstdFlags)
	return
}
