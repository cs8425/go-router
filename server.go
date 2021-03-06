package main

import (
	"net/http"
	"flag"
	"io/ioutil"
//	"log"
	"net"
	"os"
	"path"
	"sync"
//	"time"
	"errors"

	"plugin"

	"./lib/web"
	"./lib/tool"
)

var conf = flag.String("c", "config.json", "config file")
var bind = flag.String("l", ":8080", "bind port")
var verb = flag.Int("v", 6, "log verbosity")

var (
	// VERSION is injected by buildflags
	VERSION = "DEBUG"
)


type plug struct {
	PluginName  *string
	Version     *string
	loadFunc    func( string, string, web.Session, *web.Mux, *web.Log ) ( http.Handler, error )
	stopFunc    func( bool ) ( bool, error )
	logger      *web.Log
	mx          sync.Mutex
}

func main() {
/*	if VERSION == "DEBUG" {
		tool.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}*/
	flag.Parse()

	err := parseJSONConfig(&config, *conf)
	if err != nil {
		tool.Vln(1, "parse JSON Config error:", err)
		return
	}
	tool.Vln(2, "conf:", config)

	err = parseUserJSON(config.UserDB)
	if err != nil {
		tool.Vln(1, "parse User DB error:", err)
		return
	}

	tool.Verbosity = config.Verb
	tool.Vln(2, "version:", VERSION)

	lis, err := net.Listen("tcp", config.Listen)
	if err != nil {
		tool.Vln(2, "Error listening:", err.Error())
		os.Exit(1)
	}
	defer lis.Close()

	tool.Vln(2, "listening on:", lis.Addr())
	tool.Vln(2, "verbosity:", *verb)

	sess, _ := web.NewManager()
//	sess.Maxlifetime = 30 * time.Second
//	sess.Gclifetime = 10 * time.Second
	mainsess, _ := sess.NewPlugin("/")
	go sess.SessionGC()
	defer sess.StopGC()

	mux := web.NewRootMux()
	muxapi := mux.NewSubMux("/api/")
	tool.Vln(2, "root mux:", mux)
	tool.Vln(2, "root api mux:", muxapi)

	// static file
	mux.Handle("/res/", mux.StripPrefix("/res", http.FileServer(http.Dir("res"))))

	// registering status
	RegStatsHandler(mainsess, muxapi)

	// registering power control
	mux.HandleFuncAuth(mainsess, "/power", powerHandler)

	// registering login control
	RegUserHandler(mainsess, mux)


/*
	rh := http.RedirectHandler("http://example.org", 307)
	mux.Handle("/foo", rh)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		store, err := mainsess.Start(w, r)
		if err != nil {
			fmt.Fprintf(w, "SessionStart error! <br/>\n")
		}
		store.Set("isLogin", true)
		fmt.Fprintf(w, "Welcome to the login page! %s\n", r.URL.Path)
		fmt.Fprintf(w, "Session isLogin = %v\n", store.IsLogin())
	})*/

	loaded := make([]*plug, 0)
	plugins, _ := ioutil.ReadDir("plugins/")
	for _, p := range plugins {

		pname := p.Name()
		pbase := path.Join("plugins/", pname)
		tool.Vln(2, "[load-plugin]", pname, pbase)

		lib, err := loadPlugin(pname, pbase)
		if err != nil {
			tool.Vln(2, "[load-plugin][Fail]", pname, err)
			continue
		}

		urlbase := "/" + pname + "/"
		plugsess, _ := sess.NewPlugin(*lib.PluginName + "-v" + *lib.Version + "@" + pname)
		plugmux := web.NewMux(urlbase)
		pluglog, err := web.NewPluginLogger(pname, 2)
		if err != nil{
			tool.Vln(2, "[load-plugin][create log]err", pname, err)
			pluglog = web.Default
		}
		lib.logger = pluglog

		hand, err := lib.loadFunc(pname, pbase, plugsess, plugmux, pluglog)
		if err != nil {
			tool.Vln(2, "[load-plugin][Load]err", pname, err)
			continue
		}
		if hand != nil {
			mux.Handle(urlbase, hand)
		}

		tool.Vln(2, "[load-plugin][OK]", pname)

		loaded = append(loaded, lib)
	}

	http.Serve(lis, mux)
}

func loadPlugin(pname, pbase string) (*plug, error) {
	errType := errors.New("wrong type")

	lib, err := plugin.Open(path.Join(pbase, "plugin.so"))
	if err != nil {
		tool.Vln(2, "[load-plugin]open err", pname, err)
		return nil, err
	}

	name, err := lib.Lookup("PLIGIN_NAME")
	if err != nil {
		tool.Vln(2, "[load-plugin][PLIGIN_NAME]Lookup err", pname, err)
		return nil, err
	}
	plugname, ok := name.(*string)
	if !ok {
		tool.Vln(2, "[load-plugin][PLIGIN_NAME]wrong type", pname)
		return nil, errType
	}

	ver, err := lib.Lookup("VERSION")
	if err != nil {
		tool.Vln(2, "[load-plugin][VERSION]Lookup err", pname, err)
		return nil, err
	}
	version, ok := ver.(*string)
	if !ok {
		tool.Vln(2, "[load-plugin][VERSION]wrong type", pname)
		return nil, errType
	}

	load, err := lib.Lookup("OnLoad")
	if err != nil {
		tool.Vln(2, "[load-plugin][OnLoad]Lookup err", pname, err)
		return nil, err
	}
	loadImpl, ok := load.(func( string, string, web.Session, *web.Mux, *web.Log ) ( http.Handler, error ))
	if !ok {
		tool.Vln(2, "[load-plugin][OnLoad]wrong type", pname)
		return nil, errType
	}

	stop, err := lib.Lookup("OnStop")
	if err != nil {
		tool.Vln(2, "[load-plugin][OnStop]Lookup err", pname, err)
		return nil, err
	}
	stopImpl, ok := stop.(func( bool ) ( bool, error ))
	if !ok {
		tool.Vln(2, "[load-plugin][OnStop]wrong type", pname)
		return nil, errType
	}

	ins := &plug {
		PluginName: plugname,
		Version: version,
		loadFunc: loadImpl,
		stopFunc: stopImpl,
	}

	return ins, nil
}


