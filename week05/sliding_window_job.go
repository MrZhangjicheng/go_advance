package main

import (
	"sync"
	"time"
)

type Bucket struct {
	sync.RWMutex
	Number  int
	TimeNow time.Time
}

func NewBucket() *Bucket {
	return &Bucket{
		TimeNow: time.Now(),
	}
}

func (bu *Bucket) Add() {
	bu.Lock()
	defer bu.Unlock()
	bu.Number++
}

type SlidingWindow struct {
	sync.RWMutex
	Size    int
	Buckets []*Bucket
}

func NewSlidingWindow(s int) *SlidingWindow {
	return &SlidingWindow{
		Size: size,
		Buckets: make([]*Bucket,0,size)
	}
}

func (sw *SlidingWindow) AppendBucket() {
	sw.Lock()
	defer sw.Unlock()
	sw.Buckets = append(sw.Buckets,NewBucket())
	if !(len(sw.Buckets) < sw.Size+1) {
		sw.Buckets = sw.Buckets[1:]
	}
}

// 
func (sw *SlidingWindow) GetBucket() *Bucket {
	if len(sw.Buckets) == 0 {
		sw.AppendBucket()
	}
	return sw.buckets[len(sw.Buckets)-1]
}

func (sw *SlidingWindow) RecordReqResult() {
	sw.GetBucket().Add()
}

// 启动滑动窗口
func (sw *SlidingWindow) Start() {
	go func() {
		for {
			sw.AppendBucket()
			time.Sleep(time.Millisecond * 100)
		}
	}()
}