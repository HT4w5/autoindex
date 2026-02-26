package log

import "log"

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

type SimpleLogger struct {
	Level LogLevel
}

func (l *SimpleLogger) Debugf(format string, a ...any) {
	if l.Level >= Debug {
		log.Printf("[DEBUG] "+format, a...)
	}
}

func (l *SimpleLogger) Infof(format string, a ...any) {
	if l.Level >= Info {
		log.Printf("[INFO] "+format, a...)
	}
}

func (l *SimpleLogger) Warnf(format string, a ...any) {
	if l.Level >= Warn {
		log.Printf("[WARN] "+format, a...)
	}
}

func (l *SimpleLogger) Errorf(format string, a ...any) {
	if l.Level >= Error {
		log.Printf("[ERROR] "+format, a...)
	}
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
