package tool

import (
	"bufio"
	"crypto/sha256"
	"os"
	"strings"
	"log"
)

var Verbosity int = 2
var std = log.New(os.Stderr, "", log.LstdFlags)


func SetFlags(f int) {
	std.SetFlags(f)
}
func SetOutput(f *os.File) {
	std.SetOutput(f)
}
func Vf(level int, format string, v ...interface{}) {
	if level <= Verbosity {
		std.Printf(format, v...)
	}
}
func V(level int, v ...interface{}) {
	if level <= Verbosity {
		std.Print(v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= Verbosity {
		std.Println(v...)
	}
}

func ReadLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func safeXORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}

func hashBytes(a []byte) []byte {
	shah := sha256.New()
	shah.Write(a)
	return shah.Sum([]byte(""))
}

