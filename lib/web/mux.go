package web

import (
	"net/http"
	"path"
)

type Mux struct {
	base  string
	mu    *http.ServeMux
}

func NewMux(base string) *Mux {
	url := path.Clean(path.Join("/", base) + "/")
	return &Mux {
		base: url,
		mu: http.NewServeMux(),
	}
}

func (mux *Mux) Handle(pattern string, handler http.Handler) {
	mux.mu.Handle(mux.base + pattern, handler)
}

func (mux *Mux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.mu.HandleFunc(mux.base + pattern, handler)
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.mu.ServeHTTP(w, r)
}


func (mux *Mux) StripPrefix(prefix string, h http.Handler) http.Handler {
	return http.StripPrefix(mux.base + prefix, h)
}


func (mux *Mux) HandleAuth(sess *Manager, pattern string, handler http.Handler) {
	mux.HandleFuncAuth(sess, pattern, handler.ServeHTTP)
/*	mux.mu.HandleFunc(mux.base + pattern, func(w http.ResponseWriter, r *http.Request) {
		if true {
			handler.ServeHTTP(w,r)
		} else {
			http.Error(w, "Forbidden", 403)
		}
	})*/
}

func (mux *Mux) HandleFuncAuth(sess *Manager, pattern string, handler func(http.ResponseWriter, *http.Request)) {
//	mux.HandleAuth(sess, pattern, http.HandlerFunc(handler))
/*	mux.mu.HandleFunc(mux.base + pattern, func(w http.ResponseWriter, r *http.Request) {
		if true {
			handler(w,r)
		} else {
			http.Error(w, "Forbidden", 403)
		}
	})*/
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if true {
			handler(w,r)
		} else {
			http.Error(w, "Forbidden", 403)
		}
	})
}

