package log

import (
	"golang.org/x/time/rate"
	"time"
)

type TokenLimit struct {
	limiter   *rate.Limiter
}

func NewTokenLimit(token float64, burst int) *TokenLimit{
	limiter := rate.NewLimiter(rate.Limit(token), burst)
	return &TokenLimit{limiter: limiter}
}

func (l *TokenLimit) GetToken(num int) bool {
	return l.limiter.AllowN(time.Now(), num)
}
