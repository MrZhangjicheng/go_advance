package common

import (
	conf "dc/configs"
	serializer "dc/internal/serializer/common"
	"encoding/json"
	"fmt"

	"gopkg.in/go-playground/validator.v8"
	"ksogit.kingsoft.net/kdocs_backend_component/keystone/util"
)

// ErrorResponse 返回错误消息
func ErrorResponse(err error) serializer.Response {
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := conf.T(fmt.Sprintf("Field.%s", e.Field))
			tag := conf.T(fmt.Sprintf("Tag.Valid.%s", e.Tag))
			unit := "" // 单位
			if _, ok := e.Value.(string); ok {
				tags := []string{"cmin", "cmax"}
				if ok, _ := util.InArray(e.Tag, tags); ok {
					unit = "个字符(中文占两个字符)"
				}
			}
			return serializer.Response{
				Code:   serializer.CodeParamErr,
				Result: fmt.Sprintf("%s%s%s%s", field, tag, e.Param, unit),
			}
		}
	}
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return serializer.BindErr(err)
	}

	return serializer.Error(serializer.CodeParamErr, "JSON类型不匹配", err)
}
