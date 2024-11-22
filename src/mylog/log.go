package mylog

import (
	"io"
	"log"
)

type Level uint

const (
	FATAL Level = iota
	ERROR
	WARN
	INFO
	DEBUG
)

type Logger struct {
	logger      *log.Logger
	fatalLogger *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger
	level       Level
	name        string
}

func NewLogger(out io.Writer, name string, level Level) (l *Logger) {
	l = &Logger{
		level: level,

		logger:      log.New(out, "", 3),
		fatalLogger: log.New(out, "", 3),
		errorLogger: log.New(out, "", 3),
		warnLogger:  log.New(out, "", 3),
		infoLogger:  log.New(out, "", 3),
		debugLogger: log.New(out, "", 3),
	}

	l.SetName(name)
	return

}

func parsePrefix(name string, level string) string {
	return level + " " + name + " "
}

func (l Logger) GetLevel() Level {
	return l.level
}

func (l *Logger) SetLevel(lvl Level) {
	l.level = lvl
}

func (l *Logger) SetName(name string) {
	l.name = name
	l.logger.SetPrefix(parsePrefix(name, "LOG  "))
	l.fatalLogger.SetPrefix(parsePrefix(name, "FATAL"))
	l.errorLogger.SetPrefix(parsePrefix(name, "ERROR"))
	l.warnLogger.SetPrefix(parsePrefix(name, "WARN "))
	l.infoLogger.SetPrefix(parsePrefix(name, "INFO "))
	l.debugLogger.SetPrefix(parsePrefix(name, "DEBUG"))
}

// log functions

func (l Logger) Printf(format string, v ...any) {
	l.logger.Printf(format, v...)
}

func (l Logger) Println(v ...any) {
	l.logger.Println(v...)
}

func (l Logger) Fatal(v ...any) {
	l.fatalLogger.Fatal(v...)
}

func (l Logger) Fatalf(format string, v ...any) {
	l.fatalLogger.Fatalf(format, v...)
}

func (l Logger) Error(v ...any) {
	l.errorLogger.Println(v...)
}

func (l Logger) Errorf(format string, v ...any) {
	l.errorLogger.Printf(format, v...)
}

func (l Logger) Warn(v ...any) {
	l.warnLogger.Println(v...)
}

func (l Logger) Warnf(format string, v ...any) {
	l.warnLogger.Printf(format, v...)
}

func (l Logger) Info(v ...any) {
	if l.level < INFO {
		return
	}
	l.infoLogger.Println(v...)
}

func (l Logger) Infof(format string, v ...any) {
	if l.level < INFO {
		return
	}
	l.infoLogger.Printf(format, v...)
}

func (l Logger) Debug(v ...any) {
	if l.level < DEBUG {
		return
	}
	l.debugLogger.Println(v...)
}

func (l Logger) Debugf(format string, v ...any) {
	if l.level < DEBUG {
		return
	}
	l.debugLogger.Printf(format, v...)
}
