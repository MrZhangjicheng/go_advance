package log

//Handler writes logs to somewhere
type Handler interface {
	// TODO
	// `Write` only used in `IHandleIOWriteThread.AsyncWrite` for now.
	// so consider it as it's input parameter may be a more better choice.
	// `p []byte` in `Write` is too low to expose...
	Write(p []byte) (n int, err error)
	Close() error
	AsyncWrite(log *LogInstance)
	SetWriteIOThread(th IHandleIOWriteThread)
	SetFormatter(fmt Formatter)
	Clone() Handler
}
