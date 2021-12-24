package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "github.com/json-iterator/go"
)

const (
	keyFileNo = "file"
	keyTime   = "time"
	keyMsg    = "msg"
	keyLevel  = "level"
)

type LogInstance struct {
	Flag  int    `json:"-"`
	Level string `json:"level"`
	File  string `json:"file"`
	Time  string `json:"time"`
	Msg   string `json:"msg"`
	KV    Fields `json:"-"`
}

type Formatter interface {
	Format(writeTobuff *bytes.Buffer, l *LogInstance) (*bytes.Buffer, error)
}

type JSONFormatter struct{}

func (j *JSONFormatter) Format(writeTobuff *bytes.Buffer,
	l *LogInstance) (*bytes.Buffer, error) {

	var err error

	if len(l.KV) > 0 {
		l.KV[keyFileNo] = l.File
		l.KV[keyTime] = l.Time
		l.KV[keyLevel] = l.Level
		l.KV[keyMsg] = l.Msg
		err = json.NewEncoder(writeTobuff).Encode(l.KV)
	} else {
		err = json.NewEncoder(writeTobuff).Encode(l)
	}

	if err != nil {
		return nil,
			fmt.Errorf("[Log-Formatter] Failed marshal to JSON, %v", err)
	}

	return writeTobuff, nil
}

// https://jsoniter.com/index.cn.html
// type JsoniterFormatter struct{}

// func (j *JsoniterFormatter) Format(writeTobuff *bytes.Buffer,
// 	l *LogInstance) (*bytes.Buffer, error) {

// 	enc := jsoniter.ConfigCompatibleWithStandardLibrary.NewEncoder(writeTobuff)
// 	var err error

// 	if len(l.KV) > 0 {
// 		l.KV[keyFileNo] = l.File
// 		l.KV[keyTime] = l.Time
// 		l.KV[keyLevel] = l.Level
// 		l.KV[keyMsg] = l.Msg
// 		err = enc.Encode(l.KV)
// 	} else {
// 		err = enc.Encode(l)
// 	}

// 	return writeTobuff, err
// }

type TxtLineFormatter struct {
}

func (_ *TxtLineFormatter) Format(writeTobuff *bytes.Buffer,
	l *LogInstance) (*bytes.Buffer, error) {

	if l.Flag&Ltime > 0 {
		writeTobuff.WriteString(l.Time)
		writeTobuff.WriteString(FieldSplit)
	}

	if l.Flag&Llevel > 0 {
		writeTobuff.WriteString(l.Level)
		writeTobuff.WriteString(FieldSplit)
	}

	if l.Flag&Lfile > 0 {
		writeTobuff.WriteString(l.File)
		writeTobuff.WriteString(FieldSplit)
	}

	writeTobuff.WriteString(l.Msg)
	if l.Msg[len(l.Msg)-1] != '\n' {
		writeTobuff.WriteByte('\n')
	}

	return writeTobuff, nil
}
