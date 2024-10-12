package log

import (
	"fmt"
	"io"
	syslog "log"
	"os"
)

const (
	red    = "\033[31m"
	yellow = "\033[33m"
	rest   = "\033[0m"
)

var (
	defaultLogger = New(os.Stdout, "", "")
	warnLogger    = New(os.Stdout, yellow, rest)
	errorLogger   = New(os.Stderr, red, rest)
)

func New(out *os.File, prefix string, suffix string) *syslog.Logger {
	return syslog.New(&Writer{out: out, suffix: suffix}, prefix, syslog.LstdFlags)
}

type Writer struct {
	out    io.Writer
	suffix string
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.suffix != "" {
		p = fmt.Append(p, w.suffix)
	}
	return w.out.Write(p)
}

func Info(format string, v ...any) {
	defaultLogger.Printf(format, v...)
}

func Warn(format string, v ...any) {
	warnLogger.Printf(format, v...)
}

func Error(format string, v ...any) {
	errorLogger.Printf(format, v...)
}

func Panic(format string, v ...any) {
	errorLogger.Printf(format, v...)
	os.Exit(1)
}

func InfoLogger() *syslog.Logger {
	return defaultLogger
}

func ErrorLogger() *syslog.Logger {
	return errorLogger
}
