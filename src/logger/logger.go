package logger

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

const TimeFormat = "2006-01-02 15:04:05"

var LOG *Logger = NewLogger(DefaultHandler)

var LevelName [6]string = [6]string{"Trace", "Debug", "Info", "Warn", "Error", "Fatal"}

type Handler interface {
	Write(buffer []byte) (n int, err error)
	Close() error
}

type Logger struct {
	mutex   sync.Mutex
	hanlder Handler
	level   int
	flag    int
	stop    chan bool
	msg     chan []byte
	buffers [][]byte
}

func NewLogger(handler Handler) *Logger {
	l := new(Logger)
	l.hanlder = handler
	l.stop = make(chan bool)
	l.msg = make(chan []byte, 1024)
	l.buffers = make([][]byte, 0, 16)
	l.level = LevelInfo
	go l.run()
	return l
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}

func (l *Logger) SetHandler(handler Handler) {
	l.hanlder = handler
}

func (l *Logger) run() {
	for {
		select {
		case buffer := <-l.msg:
			l.hanlder.Write(buffer)
			l.pushBuf(buffer)
		case <-l.stop:
			l.hanlder.Close()
		}
	}
}

func (l *Logger) Close() {
	close(l.stop)
}

func (l *Logger) popBuf() []byte {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if len(l.buffers) == 0 {
		buffer := make([]byte, 0, 1024)
		return buffer
	} else {
		buffer := l.buffers[len(l.buffers)-1]
		l.buffers = l.buffers[0 : len(l.buffers)-1]
		return buffer
	}
}

func (l *Logger) pushBuf(buffer []byte) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.buffers) < 1024 {
		buffer = buffer[0:0]
		l.buffers = append(l.buffers, buffer)
	}

}

func (l *Logger) Output(level int, format string, v ...interface{}) {
	if l.level > level {
		return
	}

	buf := l.popBuf()
	now := time.Now().Format(TimeFormat)
	buf = append(buf, '[')
	buf = append(buf, now...)
	buf = append(buf, "] "...)
	buf = append(buf, '[')
	buf = append(buf, LevelName[level]...)
	buf = append(buf, "] "...)

	_, file, line, _ := runtime.Caller(2)

	pos := strings.LastIndex(file, "/")
	filename := file
	if pos > 0 {
		filename = file[pos+1:]
	}
	buf = append(buf, '[')
	buf = append(buf, filename...)
	buf = append(buf, "]("...)
	buf = append(buf, fmt.Sprintf("%d", line)...)
	buf = append(buf, ") "...)
	s := fmt.Sprintf(format, v...)

	buf = append(buf, s...)

	if s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.msg <- buf
}

func (l *Logger) Trace(format string, v ...interface{}) {
	l.Output(LevelTrace, format, v...)
}
func (l *Logger) Debug(format string, v ...interface{}) {
	l.Output(LevelDebug, format, v...)
}
func (l *Logger) Info(format string, v ...interface{}) {
	l.Output(LevelInfo, format, v...)
}
func (l *Logger) Warn(format string, v ...interface{}) {
	l.Output(LevelWarn, format, v...)
}
func (l *Logger) Error(format string, v ...interface{}) {
	l.Output(LevelError, format, v...)
}
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.Output(LevelFatal, format, v...)
}
