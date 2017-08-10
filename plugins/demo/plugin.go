package main

import (
	"net/http"

	"fmt"
	"../../lib/web"
)


var VERSION = "1.0.1"
var PLIGIN_NAME = "demo"

var l *web.Log

func OnLoad( plugin_name, plugin_base_path string, sess web.Session, mux *web.Mux, logger *web.Log ) (hand http.Handler, err error) {

	fmt.Println("[onLoad]", PLIGIN_NAME, plugin_name, plugin_base_path)

	l = logger
	l.Vln(2, "[onLoad]", PLIGIN_NAME, plugin_name, plugin_base_path)

	mux.Handle("/res/", mux.StripPrefix("/res", http.FileServer(http.Dir(plugin_base_path + "/res"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		store, err := sess.Start(w, r)
		fmt.Fprintf(w, "Welcome to the plugin page! %s-v%s %s <br/>\n", PLIGIN_NAME, VERSION, r.URL.Path)
		if err != nil {
			fmt.Fprintf(w, "SessionStart error! <br/>\n")
		}
		test, ok := store.Get("test")
		if !ok {
			store.Set("test", 1)
		} else {
			store.Set("test", test.(int)+1)
		}
		fmt.Fprintf(w, "Session test, ok = %v, %v <br/>\n", test, ok)
		fmt.Fprintf(w, "Session isLogin = %v <br/>\n", store.IsLogin())
	})

	return mux, nil
}

func OnStop( cleanup bool ) (ok bool, err error) {
	fmt.Println("[OnStop]", PLIGIN_NAME, cleanup)
	l.Vln(2, "[OnStop]", PLIGIN_NAME, cleanup)
	return true, nil
}

func main() {
	// nop
}

