package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"os"

	"./lib/web"
	"./lib/tool"
)

type User struct {
	Login     string `json:"login"`
	Pass      string `json:"pass"`
	Salt      string `json:"salt"`
	Sup       bool   `json:"sup"`
	Name      string `json:"name"`
	Remark    string `json:"remark"`
}

var users []User

func parseUserJSON(path string) error {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(&users)
}

func RegUserHandler( sess web.Session, mux *web.Mux ) {

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		store, err := sess.Start(w, r)
		if err != nil {
			http.Error(w, "Session error", 500)
		}

		if store.IsLogin() {
			http.Redirect(w, r, "/", 302)
		}

		switch r.Method {
		case "GET":
			fmt.Fprintf(w, "login page: %s\n", r.URL.Path)
		case "POST":
			r.ParseForm()
			user := r.Form.Get("user")
			passwd := r.Form.Get("passwd")
			tool.Vln(3, "[Login]", user, passwd)

		default:
			http.Error(w, "Method Not Allowed", 405)
		}
	})

}



func userJSONHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
}



