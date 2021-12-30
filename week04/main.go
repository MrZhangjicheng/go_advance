package main

import (
	"time"
)

//
type Bucket struct {
	Number int // 该桶的个数
}

// EndTime - StartTime = 10s,每个桶为1s
type SlidingWindow struct {
	Total     int            // 整个窗口的数量
	Buckets   map[int]Bucket // 以时间戳为key,方便定位对应的桶
	LenBucket int            // 桶的数量,max为10
	StartTime time.Time      // 最开始桶的时间
	EndTime   time.Time      // 最后面桶的时间,方便窗口滑动
}

func NewSlidingWindow() *SlidingWindow {
	return &SlidingWindow{
		Total:     0,
		Buckets:   make(map[int]Bucket),
		LenBucket: 0,
	}
}

func (sw *SlidingWindow) Add(number int) {
	// 先检查当前时间
	time.Time
}
