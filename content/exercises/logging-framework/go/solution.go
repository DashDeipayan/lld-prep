package main

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

func (l Level) String() string { panic("TODO") }

type Formatter interface {
	Format(logger string, lvl Level, msg string) string
}

type Handler interface {
	Handle(logger string, lvl Level, msg string) error
}

type SimpleFormatter struct{}

func (SimpleFormatter) Format(logger string, lvl Level, msg string) string { panic("TODO") }

type JSONFormatter struct{}

func (JSONFormatter) Format(logger string, lvl Level, msg string) string { panic("TODO") }

func NewMemoryHandler(min Level, sink *[]string, f Formatter) Handler { panic("TODO") }

type Logger struct {
	// TODO
}

func NewLogger(name string, min Level) *Logger { panic("TODO") }
func (l *Logger) AddHandler(h Handler) *Logger { panic("TODO") }
func (l *Logger) SetLevel(lvl Level)           { panic("TODO") }
func (l *Logger) Debug(msg string)             { panic("TODO") }
func (l *Logger) Info(msg string)              { panic("TODO") }
func (l *Logger) Warn(msg string)              { panic("TODO") }
func (l *Logger) Error(msg string)             { panic("TODO") }

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(logger string, lvl Level, msg string) error

func (f HandlerFunc) Handle(logger string, lvl Level, msg string) error { return f(logger, lvl, msg) }
