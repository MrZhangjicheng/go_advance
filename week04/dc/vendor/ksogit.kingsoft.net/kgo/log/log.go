package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

//log level, from low to high, more high means more serious
const (
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelBuss
)

const (
	Ltime  = 1 << iota //time format "2006/01/02 15:04:05"
	Lfile              //file.go:123
	Llevel             //[Trace|Debug|Info...]

	StdLogFlag = Ltime | Lfile | Llevel

	TimeFormat = "2006/01/02 15:04:05"
	FieldSplit = " - "
)

type Fields map[string]interface{}

var LevelString = map[string]int{
	"trace": LevelTrace,
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
	"fatal": LevelFatal,
	"buss":  LevelBuss,
}

var LevelName = [7]string{
	"TRACE", "DEBUG", "INFO", "WARN",
	"ERROR", "FATAL", "BUSS",
}

var (
	// 每条log最大允许大小（除去time\level\fileno几个字段后的msg字段最大限制）
	// 超过这个值的msg字段会截断。需要大日志引入包后，可以直接改这个值。
	MaxBytesPerLog = 1024 * 3

	// 假如每条日志3k+, 乘以这里chan长度，相当于chan满时总大内存占用。
	defaultIothreadChanBufferLength = _4k

	//----------- 全局对像 ------------------------
	//外部实现IOThread的话，请调用 GlobalLogInstenceBuffer.Put(hw.Log)
	GlobalLogInstenceBuffer *sync.Pool
	globalWriteThread       *HandleIOWriteThread
	globalTxtLineFormatter  = new(TxtLineFormatter)
)

func init() {
	GlobalLogInstenceBuffer = &sync.Pool{
		New: func() interface{} {
			return new(LogInstance)
		}}

	globalWriteThread = NewHandleIOWriteThread("globalLogIOThread", defaultIothreadChanBufferLength) // default length=4K
}

type Logger struct {
	level int
	flag  int

	handlers []Handler

	kv Fields
}

//new a logger with specified handler and flag
func NewLogger(handler Handler, flag int) *Logger {

	var l = new(Logger)

	l.level = LevelInfo

	l.handlers = make([]Handler, 1)
	l.handlers[0] = handler

	l.flag = flag
	l.kv = make(Fields, 5)

	return l
}

// func New() name alais
var New = NewLogger

//new a default logger with specified handler and flag: Ltime|Lfile|Llevel
func NewDefaultLogger(handler Handler) *Logger {
	return NewLogger(handler, Ltime|Lfile|Llevel)
}

func newStdHandler() *StreamHandler {
	h, _ := NewStreamHandler(os.Stdout)
	return h
}

var std = NewDefaultLogger(newStdHandler())

type manager struct {
	mapper map[string]interface{}
	mu     sync.RWMutex
}

func newManager() *manager {
	m := new(manager)
	m.mapper = make(map[string]interface{})
	return m
}

func (self *manager) get(name string) *Logger {
	self.mu.Lock()
	defer self.mu.Unlock()

	l, ok := self.mapper[name]
	if ok {
		return l.(*Logger)
	} else {
		l = NewDefaultLogger(newStdHandler())
		self.mapper[name] = l
	}
	return l.(*Logger)
}

func (self *manager) close() {
	for _, v := range self.mapper {
		v.(*Logger).Close()
	}
}

var _mgr = newManager()

// like the python logging.getLogger
// return an Gloabl-logger and save in the memory
func GetLogger(name string) *Logger {
	if name == "" || name == "root" {
		return std
	}
	return _mgr.get(name)
}

func Close() {
	globalWriteThread.Close()
	// std.Close()
	_mgr.close()
}

func (l *Logger) Close() {
	if l.handlers != nil {
		for _, h := range l.handlers {
			h.Close()
		}
	}
}

//set log level, any log level less than it will not log
func (l *Logger) SetLevel(level int) {
	l.level = level
}

func (l *Logger) Level() int {
	return l.level
}

// when expect Logger has only one Handler, use this function
func (l *Logger) SetHandler(h Handler) {
	l.handlers[0] = h
}

// when expect Logger more the one Handler, use this function
func (l *Logger) AppendHandler(h Handler) {
	l.handlers = append(l.handlers, h)
}

//a low interface, maybe you can use it for your special log format
//but it may be not exported later......
func (l *Logger) Output(callDepth int, level int, format string, v ...interface{}) {
	if l.level > level {
		return
	}

	msg := fmt.Sprintf(format, v...)
	l.OutputMsg(callDepth, level, msg)
}

func (l *Logger) OutputMsg(callDepth int, level int, msg string) {
	if l.level > level {
		return
	}

	var file_line, now, slevel string

	if l.flag&Ltime > 0 {
		now = time.Now().Format(TimeFormat)
	}

	if l.flag&Llevel > 0 {
		slevel = LevelName[level]
	}

	if l.flag&Lfile > 0 {
		_, file, line, ok := runtime.Caller(callDepth)
		if !ok {
			file = "???"
			line = 0
		} else {
			v := strings.Split(file, "/")
			idx := len(v) - 3
			if idx < 0 {
				idx = 0
			}
			file = strings.Join(v[idx:], "/")
		}

		file_line = fmt.Sprintf("%s:[%d]", file, line)
	}

	if len(msg) > MaxBytesPerLog {
		// MAX_BYTES_PER_LOG 默认是3k
		// 只允许写入3K日志数据防止日志太长内存拷贝以及IO上升.
		//   -- by pengweilin 2017-03-02
		msg = fmt.Sprintf("%s... data too long, soucre-length=%d",
			msg[0:MaxBytesPerLog], len(msg))
	}

	// 行号符提成 "||"
	msg = strings.Replace(msg, "\n", "||", -1)
	// 回车符替换成 "||"
	msg = strings.Replace(msg, "\r", "||", -1)

	log := GlobalLogInstenceBuffer.Get().(*LogInstance)
	log.Flag = l.flag
	log.File = file_line
	log.Level = slevel
	log.KV = l.kv
	log.Time = now
	log.Msg = msg
	// log := LogInstance{
	// 	Flag:  l.flag,
	// 	Time:  now,
	// 	Level: slevel,
	// 	Msg:   msg,
	// 	File:  file_line,
	// 	KV:    l.kv,
	// }

	for _, h := range l.handlers {
		if h != nil {
			h.AsyncWrite(log)
		}
	}
}

//log with Trace level
func (l *Logger) Trace(format string, v ...interface{}) {
	l.Output(3, LevelTrace, format, v...)
}

//log with Debug level
func (l *Logger) Debug(format string, v ...interface{}) {
	l.Output(3, LevelDebug, format, v...)
}

//log with info level
func (l *Logger) Info(format string, v ...interface{}) {
	l.Output(3, LevelInfo, format, v...)
}

//log with warn level
func (l *Logger) Warn(format string, v ...interface{}) {
	l.Output(3, LevelWarn, format, v...)
}

//log with error level
func (l *Logger) Error(format string, v ...interface{}) {
	l.Output(3, LevelError, format, v...)
}

//log with fatal level
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.Output(3, LevelFatal, format, v...)
}

func (l *Logger) Buss(format string, v ...interface{}) {
	l.Output(3, LevelBuss, format, v...)
}

func (l *Logger) SetFormatter(f Formatter) {
	for _, h := range l.handlers {
		h.SetFormatter(f)
	}
}

func SetLevel(level int) { std.SetLevel(level) }

func SetHandler(h Handler) { std.SetHandler(h) }

func AppendHandler(h Handler) { std.AppendHandler(h) }

func SetLevelS(level string) {
	SetLevel(LevelString[strings.ToLower(level)])
}

func SetDropCallback(f DropLogCallbackFunc) {
	globalWriteThread.SetDropCallback(f)
}

func SetLimitCallback(f DropLogCallbackFunc) {
	globalWriteThread.SetLimitCallback(f)
}

func SetGlobalLimiter(token float64, burst int) {
	globalWriteThread.SetLimiter(token, burst)
}

func SetGlobalWriteThreadChanBufferLen(length int) {
	if length <= 0 {
		panic("buffer length must >0.")
		return
	}
	// TODO: maybe need a lock.
	globalWriteThread.Close()
	globalWriteThread = NewHandleIOWriteThread("globalLogIOThread", length)
}

func Trace(format string, v ...interface{}) {
	std.Output(3, LevelTrace, format, v...)
}

func Debug(format string, v ...interface{}) {
	std.Output(3, LevelDebug, format, v...)
}

func Info(format string, v ...interface{}) {
	std.Output(3, LevelInfo, format, v...)
}

func Warn(format string, v ...interface{}) {
	std.Output(3, LevelWarn, format, v...)
}

func Error(format string, v ...interface{}) {
	std.Output(3, LevelError, format, v...)
}

func Fatal(format string, v ...interface{}) {
	std.Output(3, LevelFatal, format, v...)
}

func Buss(format string, v ...interface{}) {
	std.Output(3, LevelBuss, format, v...)
}
func SetFormatter(f Formatter) {
	std.SetFormatter(f)
}

// Function alais
var WithField = std.WithField
var WithFields = std.WithFields

// func WithField(k string, v interface{}) *Logger {
// 	return std.WithField(k, v)
// }
// func WithFields(kv Fields) *Logger {
// 	return std.WithFields(kv)
// }

func StdLogger() *Logger {
	return std
}

func GetLevel() int {
	return std.level
}
