package main

import (
//	"encoding/json"
	"net/http"
	"fmt"

	"./lib/tool"
)

func powerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "power control page: %s\n", r.URL.Path)
		fmt.Fprintf(w, "power control option: poweroff, reboot\n")
	case "POST":
		r.ParseForm()
		do := r.Form.Get("do")
		tool.Vln(3, "[PowerCtrl]", do)
		fmt.Fprintf(w, "power goes: %s!\n", do)
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
}

func powerJSONHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
}



