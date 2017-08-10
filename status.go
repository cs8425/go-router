package main

import (
	"runtime"
	"expvar"

	"./lib/web"
)

type GoStats struct {
	NumCPU        int
	NumGoroutine  int
	NumCgoCall    int64

	MemTotal      int
	MemUsed       int // = MemTotal- (MemCached + MemBuffers + MemFree)
	MemFree       int
	MemBuffers    int
	MemCached     int
	SwapTotal     int
	SwapFree      int

	memStats      runtime.MemStats

// runtime.ReadMemStats(m *runtime.MemStats)
	HeapAlloc uint64
	HeapInuse uint64
	HeapObjects uint64
	StackInuse uint64
	NumGC uint32
	NextGC uint64
	LastGC uint64
}

var status GoStats

func GetGoStats() (interface{}) {
	runtime.ReadMemStats(&status.memStats)

	status.NumCPU = runtime.NumCPU()
	status.NumGoroutine = runtime.NumGoroutine()
	status.NumCgoCall = runtime.NumCgoCall()

	status.HeapAlloc = status.memStats.HeapAlloc
	status.HeapInuse = status.memStats.HeapInuse
	status.HeapObjects = status.memStats.HeapObjects
	status.StackInuse = status.memStats.StackInuse
	status.NumGC = status.memStats.NumGC
	status.NextGC = status.memStats.NextGC
	status.LastGC = status.memStats.LastGC

	return status
}

func init() {
	expvar.Publish("status", expvar.Func(GetGoStats))
}

func RegStatsHandler( sess web.Session, mux *web.Mux ) {
	mux.Handle("/status", expvar.Handler())
//	mux.HandleAuth("/status", expvar.Handler())
}

