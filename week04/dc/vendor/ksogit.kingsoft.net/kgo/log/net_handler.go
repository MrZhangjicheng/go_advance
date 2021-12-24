package log

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"ksogit.kingsoft.net/kgo/log/dsn"
)

type NetHandlerConfig struct {
	Agent   string
	Proto   string        `dsn:"network"`
	Addr    string        `dsn:"address"`
	Chan    int           `dsn:"query.chan"`
	Timeout time.Duration `dsn:"query.timeout"`
}

type NetHandler struct {
	c    *NetHandlerConfig
	fmt  Formatter
	conn net.Conn

	writeThread IHandleIOWriteThread
}

const (
	//_agentTimeout = 20 * time.Millisecond
	_defaultAgentConfig = "unixpacket:///tmp/collector_tcp.sock?timeout=100&chan=1024"
)

type NetHander struct {
}

// parseDSN parse log agent dsn.
// unixgram:///var/run/lancer/collector.sock?timeout=100&chan=1024
func parseDSN(rawdsn string) *NetHandlerConfig {
	ac := new(NetHandlerConfig)
	d, err := dsn.Parse(rawdsn)
	if err != nil {
		panic(errors.WithMessage(err, fmt.Sprintf("log: invalid dsn: %s", rawdsn)))
	}
	if _, err = d.Bind(ac); err != nil {
		panic(errors.WithMessage(err, fmt.Sprintf("log: invalid dsn: %s", rawdsn)))
	}
	return ac
}

func NewNetHandler(dc *NetHandlerConfig) (a *NetHandler) {
	if dc == nil {
		dc = parseDSN(_defaultAgentConfig)
	}

	a = &NetHandler{
		c: dc,
	}

	a.fmt = globalTxtLineFormatter

	return
}

// TODO simple net log protocol, `|len(4B)|body...|`
// You can call `SetWriteIOThread` to write you own protocol.
func (h *NetHandler) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("empty []byte to Write")
	}

	// TODO reuse such buffer to avoid memory allocation frequently...
	buffer := make([]byte, 4, 4+len(b))

	// write head
	binary.BigEndian.PutUint32(buffer, uint32(len(b)))

	// append body
	buffer = append(buffer, b...)

	if h.conn == nil {
		if h.conn, err = net.DialTimeout(h.c.Proto, h.c.Addr, time.Duration(h.c.Timeout)); err != nil {
			fmt.Printf("net.DialTimeout(%s:%s) error(%v)\n", h.c.Proto, h.c.Addr, err)
			return 0, fmt.Errorf("connect remote error:%v", err)
		}
	}

	if h.conn != nil {
		if _, err = h.conn.Write(buffer); err != nil {
			fmt.Printf("conn.Write(%d bytes) error(%v)\n", len(buffer), err)
			_ = h.conn.Close()
		} else {
			//fmt.Printf("conn.Write(%d bytes) success\n", len(buffer))
			// only succeed reset buffer, let conn reconnect.
			//buffer.Reset()
		}
	}

	return len(buffer), nil
}

func (h *NetHandler) AsyncWrite(log *LogInstance) {
	if h.writeThread != nil {
		h.writeThread.AsyncWrite(h, h.fmt, log)
	} else {
		globalWriteThread.AsyncWrite(h, h.fmt, log)
	}
}

func (h *NetHandler) SetWriteIOThread(th IHandleIOWriteThread) {
	h.writeThread = th
}

func (h *NetHandler) Close() error {
	if h.writeThread != nil {
		h.writeThread.Close()
	}

	if h.conn != nil {
		_ = h.conn.Close()
	}
	return nil
}

func (h *NetHandler) SetFormatter(fmt Formatter) {
	h.fmt = fmt
}

func (h *NetHandler) Clone() Handler {
	return nil
}
