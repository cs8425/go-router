package main

import (
	"encoding/json"
	"net/http"
	"runtime"
)

type GoStats struct {
	NumCPU        int
	NumGoroutine  int
	NumCgoCall    int64

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

func GetGoStats() (GoStats) {
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

func goStatsHandler(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintf(w, "this page need login! %s\n", r.URL.Path)

	st := GetGoStats()

	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(st)
	case "POST":
		http.Error(w, "Not yet!", 501)
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
}

func goStatsJSONHandler(w http.ResponseWriter, r *http.Request) {
	st := GetGoStats()

	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(st)
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
}



