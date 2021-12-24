package common

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code   int64       `json:"code"`
	Result string      `json:"result"`
	Data   interface{} `json:"data,omitempty"`
}

const (
	//CodeParamErr 参数错误
	CodeParamErr = -1
	//CodeDBOperateErr 数据库操作错误
	CodeDBOperateErr = -2
)

// Error 错误消息
func Error(code int64, reason string, err error) Response {
	res := Response{
		Code:   code,
		Result: reason,
	}
	if gin.Mode() != gin.ReleaseMode && err != nil {
		res.Data = fmt.Sprint(err)
	}
	return res
}

// BindErr 提交的数据有误
func BindErr(err error) Response {
	return Error(CodeParamErr, "JSON类型不匹配", err)
}

// DBOperateErr 数据库保存错误
func DBOperateErr(err error) Response {
	return Error(CodeDBOperateErr, "内部错误", err)
}
