package log

var globalJsonFormatter = new(JSONFormatter)

func (l *Logger) clone() *Logger {

	ll := new(Logger)

	ll.kv = make(Fields, len(l.kv))
	ll.level = l.level
	ll.flag = l.flag

	ll.handlers = make([]Handler, len(l.handlers))
	for i, h := range l.handlers {
		ll.handlers[i] = h.Clone()
		// 在这里强行换成Json-Formatter
		ll.handlers[i].SetFormatter(globalJsonFormatter)
	}

	for k, v := range l.kv {
		ll.kv[k] = v
	}
	return ll
}

func (l *Logger) WithField(k string, v interface{}) *Logger {
	ll := l.clone()

	ll.kv[k] = v
	return ll
}

func (l *Logger) WithFields(kv Fields) *Logger {
	ll := l.clone()

	for k, v := range kv {
		ll.kv[k] = v
	}
	return ll
}
