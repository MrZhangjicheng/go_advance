/*
整个log包，只有一个IO对象:
  如非必要，不用配置多个写线程，默认只有1个IO线程。
  如有多个handler（如需要写到 socket、多个文件)时,
  再配置多个IO线程。

*/
package log

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var stdErrLog = log.New(os.Stderr, "[/go/log] ", log.Ldate|log.Ltime|log.Lshortfile)

type DropLogCallbackFunc func(l *LogInstance, drop int)

type IHandleIOWriteThread interface {
	AsyncWrite(h Handler, fmt Formatter, log *LogInstance)

	// 当调用AsyncWrite异步写的chan满了，会直接丢弃log；
	// 丢弃前，通过DropLogCallbackFunc 回调一次，告诉上层应用。
	SetDropCallback(f DropLogCallbackFunc)
	SetLimitCallback(f DropLogCallbackFunc)
	SetLimiter(token float64, burst int)

	Close()
}

const (
	MaxWaitTimeOnExit = time.Second * 10

	_8k = 8192
	_4k = 4096 //假定文件系统block size=4k
)

type hdlrWriter struct {
	Handler Handler
	Fmt     Formatter
	Log     *LogInstance
}

type HandleIOWriteThread struct {
	name   string
	closed bool
	quit   chan bool

	handlerWriterChan   chan *hdlrWriter // 一个IO线程处理多个handler的写
	handlerWriterBuffer *sync.Pool

	writeBuffer *bytes.Buffer

	dropCnt     int64
	writeCnt    int64
	asyncSumCnt int64
	errPrintCnt int64
	limitCnt    int64

	wg                   sync.WaitGroup
	dropLogCallbackFunc  DropLogCallbackFunc
	limitLogCallbackFunc DropLogCallbackFunc

	limiter     *TokenLimit
}

func NewHandleIOWriteThread(name string, chanLength int) *HandleIOWriteThread {

	self := new(HandleIOWriteThread)

	self.name = name
	self.quit = make(chan bool, 10)
	self.handlerWriterChan = make(chan *hdlrWriter, chanLength)

	self.handlerWriterBuffer = &sync.Pool{
		New: func() interface{} {
			return new(hdlrWriter)
		},
	}

	// use 8k buffer in memory, linux filesys block was 4k
	self.writeBuffer = bytes.NewBuffer(make([]byte, 0, _8k))

	self.wg.Add(1)
	go self.run()
	return self
}

func (self *HandleIOWriteThread) SetDropCallback(f DropLogCallbackFunc) {
	self.dropLogCallbackFunc = f
}

func (self *HandleIOWriteThread) SetLimitCallback(f DropLogCallbackFunc) {
	self.limitLogCallbackFunc = f
}

func (self *HandleIOWriteThread) SetLimiter(token float64, burst int) {
	self.limiter = NewTokenLimit(token, burst)
}

func (self *HandleIOWriteThread) AsyncWrite(
	h Handler, fmt Formatter, log *LogInstance) {

	if h == nil || fmt == nil {
		return
	}

	atomic.AddInt64(&self.asyncSumCnt, 1)

	if !self.onPreWrite(log) {
		return
	}

	hw := self.handlerWriterBuffer.Get().(*hdlrWriter)
	hw.Handler = h
	hw.Fmt = fmt
	hw.Log = log

	select {
	case self.handlerWriterChan <- hw:
		return
	default:
		// Note: 当handlerWriterChan满了时，只能丢弃日志，
		//    问题在于怎么通知开发人员，丢日志了，
		//    初步想法可以通过普罗米修斯这类的数据收集，进行告警。
		//    丢日志原因有很多，可能硬盘介质写速度太慢，或满了。
		//    如果是网络发送，也会有慢的时候。
		atomic.AddInt64(&self.dropCnt, 1)
		if self.dropLogCallbackFunc != nil {
			self.dropLogCallbackFunc(log, 1)
		}
	}
}

// 限流判定
func (self *HandleIOWriteThread) onPreWrite(log *LogInstance) bool {
	if self.limiter != nil {
		if !self.limiter.GetToken(1) {
			atomic.AddInt64(&self.limitCnt, 1)
			if self.limitLogCallbackFunc != nil {
				self.limitLogCallbackFunc(log, 1)
			}
			return false
		}
	}
	return true
}

// 错误太多时，按一定概率输出，减少性能损失
func (self *HandleIOWriteThread) errRate() bool {
	atomic.AddInt64(&self.errPrintCnt, 1)
	output := false
	switch {
	case self.errPrintCnt <= 100:
		output = true
	case self.errPrintCnt <= 500:
		// 10% 概率输出
		output = bool(rand.Int31n(100) <= 10)
	case self.errPrintCnt > 1000:
		// 1% 概率输出
		output = bool(rand.Int31n(100) <= 1)
	}
	return output
}

func (self *HandleIOWriteThread) doFormat(hw *hdlrWriter,
	buff *bytes.Buffer) {

	defer func() {
		GlobalLogInstenceBuffer.Put(hw.Log)
		self.handlerWriterBuffer.Put(hw)

		if r := recover(); r != nil && self.errRate() {
			stdErrLog.Printf("Log[%s] doFormat recover=[%v]\n",
				self.name, r)
		}
	}()

	if hw.Fmt == nil {
		hw.Fmt = globalTxtLineFormatter
	}

	if _, e := hw.Fmt.Format(buff, hw.Log); e != nil {
		//TODO: format出错，怎么办？
		if hw.Fmt != globalTxtLineFormatter {
			// TxtLineFormatter 不会返回出错
			hw.Fmt.Format(buff, hw.Log)
		} else {
			self.dropCnt += 1
			if self.errRate() {
				stdErrLog.Printf("Log[%s] drop-logs[%d] format err=[%v]\n",
					self.name, self.dropCnt, e)
			}
			return
		}
	}

	self.writeCnt += 1
}

func (self *HandleIOWriteThread) doWrite(hw *hdlrWriter) {
	if hw == nil {
		return
	}

	pBuff := self.writeBuffer //default 8k
	var next *hdlrWriter

	for hw != nil {
		self.doFormat(hw, pBuff)
		if pBuff.Len() >= _4k {
			// NOTE: Handler自行处理 Write错误
			hw.Handler.Write(pBuff.Bytes())
			pBuff.Reset()
		}
		if !self.closed && len(self.handlerWriterChan) > 0 {
			next = <-self.handlerWriterChan
			if next.Handler != hw.Handler && pBuff.Len() > 0 {
				hw.Handler.Write(pBuff.Bytes())
				pBuff.Reset()
			}
			hw = next
		} else {
			break
		}
	}

	if pBuff.Len() > 0 {
		_, e := hw.Handler.Write(pBuff.Bytes())
		if e != nil && self.errRate() {
			stdErrLog.Printf("Log[%s] Write[%s] err=[%v]\n",
				self.name, string(pBuff.Bytes()), e)
		}
		pBuff.Reset()
	}
}

func (self *HandleIOWriteThread) run() {
	defer self.wg.Done()

	var stop = false
	var hw *hdlrWriter
	var quitStartTime time.Time

	for {
		select {
		case <-self.quit:
			stop = true
			quitStartTime = time.Now()
		case hw = <-self.handlerWriterChan:
			self.doWrite(hw)
		}
		if !stop {
			continue
		}

		if stop && len(self.handlerWriterChan) == 0 {
			// fmt.Fprintf(os.Stderr, "[loger] handlerWriterChan was empty. close run-thread.\n")
			return
		}

		if time.Since(quitStartTime) >= MaxWaitTimeOnExit {
			remain := len(self.handlerWriterChan)
			self.dropCnt += int64(remain)

			if remain > 0 {
				// 这句让它 100%输出到stderr.
				fmt.Fprintf(os.Stderr,
					"%s, but remain logs[%v] do not flush to IO yet.\n",
					"[kgo/log package] was Closed()", remain)

				if self.dropLogCallbackFunc != nil {
					hw = <-self.handlerWriterChan
					self.dropLogCallbackFunc(hw.Log, remain)
				}
			}
			return
		}
	} // End for-loop
}

func (self *HandleIOWriteThread) Close() {
	// TODO Lock
	if self.closed {
		return
	}
	self.closed = true
	select {
	case self.quit <- true:
		self.wg.Wait()
	default:
		return
	}
}

type Statisitc struct {
	Name     string
	Sum      int64
	WriteSum int64
	DropSum  int64
	LimitSum int64
}

func (self *HandleIOWriteThread) Stat() Statisitc {
	return Statisitc{
		Name:     fmt.Sprintf("%s:closed=%v", self.name, self.closed),
		Sum:      self.asyncSumCnt,
		WriteSum: self.writeCnt,
		DropSum:  self.dropCnt,
		LimitSum: self.limitCnt,
	}
}

// 测试用函数
func GlobalIOThreadStat() Statisitc {
	return globalWriteThread.Stat()
}
