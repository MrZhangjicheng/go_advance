package week02

import (
	"github.com/pkg/errors"
)

const (
	NoDataError string = "sql.ErrNoRows"
)

// 获得错误的元数据
func baseError() error {
	return errors.New(NoDataError)
}

// WarpError 包装错误
/*
	在真实的业务逻辑中,获取不到数据是正常的逻辑,所以遇到这种错误,我们应该直接处理
	这样在上一层调用时,先判断 err != nil, 然后就判断 返回的值是否有效
	如果在本层不处理 "sql.ErrNoRows" 这种错误，那么在上层判断 err != nil 时,还需要
	获取err的元数据，并进行判断
*/
func WarpError() (string, error) {
	var data string
	err := baseError()
	if err == nil {
		return data, nil
	} else {
		switch err.Error() {
		case NoDataError:
			return "", nil
		default:
			return "", errors.Wrap(err, "other error")
		}

	}
}
