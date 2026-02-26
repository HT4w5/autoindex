package log

import (
	"fmt"
	"os"
	"time"

	"github.com/valyala/bytebufferpool"
)

type Logger interface {
	Debugf(format string, a ...any)
	Infof(format string, a ...any)
	Warnf(format string, a ...any)
	Errorf(format string, a ...any)
}

type LogLevel int

const (
	None LogLevel = iota - 1
	Error
	Warn
	Info
	Debug
)

var (
	errorTagBytes = []byte(" [ERROR] ")
	warnTagBytes  = []byte(" [WARN] ")
	infoTagBytes  = []byte(" [INFO] ")
	debugTagBytes = []byte(" [DEBUG] ")
)

type SimpleLogger struct {
	Level LogLevel
}

func (l *SimpleLogger) Debugf(format string, a ...any) {
	if l.Level >= Debug {
		logf(debugTagBytes, format, a...)
	}
}

func (l *SimpleLogger) Infof(format string, a ...any) {
	if l.Level >= Info {
		logf(infoTagBytes, format, a...)
	}
}

func (l *SimpleLogger) Warnf(format string, a ...any) {
	if l.Level >= Warn {
		logf(warnTagBytes, format, a...)
	}
}

func (l *SimpleLogger) Errorf(format string, a ...any) {
	if l.Level >= Error {
		logf(errorTagBytes, format, a...)
	}
}

func logf(levelTag []byte, format string, a ...any) {
	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)

	bb.WriteString(time.Now().Format(time.RFC3339))
	bb.Write(levelTag)
	fmt.Fprintf(bb, format, a...)
	bb.WriteByte('\n')

	bb.WriteTo(os.Stdout)
}

type DiscardLogger struct {
}

func (l *DiscardLogger) Debugf(format string, a ...any) {
}

func (l *DiscardLogger) Infof(format string, a ...any) {
}

func (l *DiscardLogger) Warnf(format string, a ...any) {
}

func (l *DiscardLogger) Errorf(format string, a ...any) {
}
