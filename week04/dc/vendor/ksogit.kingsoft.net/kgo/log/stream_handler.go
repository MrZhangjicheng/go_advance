package log

import "io"

//StreamHandler writes logs to a specified io Writer, maybe stdout, stderr, etc...
type StreamHandler struct {
	w           io.Writer
	writeThread IHandleIOWriteThread
	fmt         Formatter
}

func NewStreamHandler(w io.Writer) (*StreamHandler, error) {
	h := new(StreamHandler)

	h.w = w
	h.fmt = globalTxtLineFormatter
	h.writeThread = nil

	return h, nil
}

func (h *StreamHandler) Clone() Handler {
	c := new(StreamHandler)
	c.w = h.w
	c.fmt = h.fmt
	c.writeThread = h.writeThread
	return c
}

func (h *StreamHandler) AsyncWrite(log *LogInstance) {
	if h.writeThread != nil {
		h.writeThread.AsyncWrite(h, h.fmt, log)
	} else {
		globalWriteThread.AsyncWrite(h, h.fmt, log)
	}
}

func (h *StreamHandler) Write(b []byte) (n int, err error) {
	return h.w.Write(b)
}

func (h *StreamHandler) SetWriteIOThread(th IHandleIOWriteThread) {
	h.writeThread = th
}

func (h *StreamHandler) SetFormatter(fmt Formatter) {
	h.fmt = fmt
}

func (h *StreamHandler) Close() error {
	if h.writeThread != nil {
		h.writeThread.Close()
	}
	return nil
}
