package kms

import "errors"

var ErrUnsupportAPI = errors.New("this API is not supported")
var ErrInvalidKeyID = errors.New("Invalid keyID")
