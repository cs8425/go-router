package web

import (
	"os"
	"log"
	"path/filepath"
)

type Log struct {
	*log.Logger
	fd     *os.File
	Verb   int
}

var Default = &Log{ log.New(os.Stderr, "", log.LstdFlags), nil, 2 }

func NewPluginLogger(tag string, verb int) (*Log, error) {
	fd, err := os.OpenFile(filepath.Join("log", tag + ".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		return nil, err
	}

	vlog := &Log{ log.New(fd, "", log.LstdFlags), fd, verb }
	return vlog, nil
}

func NewLogger() *Log {
	return &Log{ log.New(os.Stderr, "", log.LstdFlags), nil, 2 }
}

func (l *Log) Close() error {
	if l.fd != nil {
		return l.fd.Close()
	}
	return nil
}

func (l *Log) Vf(level int, format string, v ...interface{}) {
	if level <= l.Verb {
		l.Printf(format, v...)
	}
}
func (l *Log) V(level int, v ...interface{}) {
	if level <= l.Verb {
		l.Print(v...)
	}
}
func (l *Log) Vln(level int, v ...interface{}) {
	if level <= l.Verb {
		l.Println(v...)
	}
}

