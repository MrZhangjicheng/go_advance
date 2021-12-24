package log

import (
	"fmt"
	"os"
	"path"
	"time"
)

//FileHandler writes log to a file.
type FileHandler struct {
	*StreamHandler //匿名继承，写少些代码
	fd             *os.File
	fileName       string
}

func NewFileHandler(fileName string) (*FileHandler, error) {
	dir := path.Dir(fileName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0777); err != nil {
			if os.ErrExist != err {
				return nil, err
			}
		}
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	h := new(FileHandler)

	h.StreamHandler, _ = NewStreamHandler(f)
	h.fd = f
	h.fileName = fileName

	return h, nil
}

func (h *FileHandler) Clone() Handler {
	c := new(FileHandler)
	c.StreamHandler = h.StreamHandler.Clone().(*StreamHandler)
	c.fd = h.fd
	c.fileName = h.fileName
	return c
}

func (h *FileHandler) Close() error {
	if h.fd != nil {
		return h.fd.Close()
	}
	return nil
}

//RotatingFileHandler writes log a file, if file size exceeds maxBytes,
//it will backup current file and open a new one.
//max backup file number is set by backupCount, it will delete oldest if backups too many.
type RotatingFileHandler struct {
	*FileHandler

	maxBytes    int
	backupCount int
}

func NewRotatingFileHandler(fileName string, maxBytes int, backupCount int) (*RotatingFileHandler, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("invalid max bytes")
	}

	fh, err := NewFileHandler(fileName)
	if err != nil {
		return nil, err
	}

	h := new(RotatingFileHandler)
	h.FileHandler = fh
	h.fileName = fileName
	h.maxBytes = maxBytes
	h.backupCount = backupCount

	return h, nil
}

func (h *RotatingFileHandler) Clone() Handler {
	c := new(RotatingFileHandler)
	c.FileHandler = h.FileHandler.Clone().(*FileHandler)
	c.maxBytes = h.maxBytes
	c.backupCount = h.backupCount
	return c
}

func (h *RotatingFileHandler) AsyncWrite(log *LogInstance) {
	if h.writeThread != nil {
		h.writeThread.AsyncWrite(h, h.fmt, log)
	} else {
		globalWriteThread.AsyncWrite(h, h.fmt, log)
	}
}

func (h *RotatingFileHandler) SetWriteIOThread(th IHandleIOWriteThread) {
	h.writeThread = th
}

func (h *RotatingFileHandler) SetFormatter(fmt Formatter) {
	h.fmt = fmt
}

func (h *RotatingFileHandler) Write(p []byte) (n int, err error) {
	h.doRollover()
	return h.fd.Write(p)
}

func (h *RotatingFileHandler) doRollover() {
	f, err := h.fd.Stat()
	if err != nil {
		return
	}

	if h.maxBytes <= 0 {
		return
	} else if f.Size() < int64(h.maxBytes) {
		return
	}

	if h.backupCount > 0 {
		h.fd.Close()

		for i := h.backupCount - 1; i > 0; i-- {
			sfn := fmt.Sprintf("%s.%d", h.fileName, i)
			dfn := fmt.Sprintf("%s.%d", h.fileName, i+1)

			os.Rename(sfn, dfn)
		}

		dfn := fmt.Sprintf("%s.1", h.fileName)
		os.Rename(h.fileName, dfn)

		h.fd, _ = os.OpenFile(h.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
}

func (h *RotatingFileHandler) Close() error {
	if h.writeThread != nil {
		h.writeThread.Close()
	}
	if h.fd != nil {
		return h.fd.Close()
	}
	return nil
}

//TimeRotatingFileHandler writes log to a file,
//it will backup current and open a new one, with a period time you sepecified.
//
//refer: http://docs.python.org/2/library/logging.handlers.html.
//same like python TimedRotatingFileHandler.
type TimeRotatingFileHandler struct {
	*FileHandler

	baseName   string
	interval   int64
	suffix     string
	rolloverAt int64
}

const (
	WhenSecond = iota
	WhenMinute
	WhenHour
	WhenDay
)

func NewTimeRotatingFileHandler(baseName string, when int8, interval int) (*TimeRotatingFileHandler, error) {
	fh, err := NewFileHandler(baseName)

	if err != nil {
		return nil, err
	}

	h := new(TimeRotatingFileHandler)

	h.FileHandler = fh
	h.baseName = baseName

	switch when {
	case WhenSecond:
		h.interval = 1
		h.suffix = "2006-01-02_15-04-05"
	case WhenMinute:
		h.interval = 60
		h.suffix = "2006-01-02_15-04"
	case WhenHour:
		h.interval = 3600
		h.suffix = "2006-01-02_15"
	case WhenDay:
		h.interval = 3600 * 24
		h.suffix = "2006-01-02"
	default:
		return nil, fmt.Errorf("invalid when_rotate: %d", when)
	}

	h.interval = h.interval * int64(interval)

	fInfo, _ := h.fd.Stat()
	h.rolloverAt = fInfo.ModTime().Unix() + h.interval

	return h, nil
}

func (h *TimeRotatingFileHandler) Clone() Handler {
	c := new(TimeRotatingFileHandler)
	c.FileHandler = h.FileHandler.Clone().(*FileHandler)
	c.baseName = h.baseName
	c.interval = h.interval
	c.suffix = h.suffix
	c.rolloverAt = h.rolloverAt
	return c
}

func (h *TimeRotatingFileHandler) doRollover() {
	//refer http://hg.python.org/cpython/file/2.7/Lib/logging/handlers.py
	now := time.Now()

	if h.rolloverAt <= now.Unix() {
		fName := h.baseName + now.Format(h.suffix)
		h.fd.Close()
		e := os.Rename(h.baseName, fName)
		if e != nil {
			panic(e)
		}

		h.fd, _ = os.OpenFile(h.baseName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

		h.rolloverAt = time.Now().Unix() + h.interval
	}
}

func (h *TimeRotatingFileHandler) Write(b []byte) (n int, err error) {
	h.doRollover()
	return h.fd.Write(b)
}
