package main

import (
	"net/http"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"sync"
	"time"
	"errors"

	"fmt"

	PL "plugin"

	"./lib/web"
)

//var conf = flag.String("c", "server.json", "config file")
var bind = flag.String("l", ":8080", "bind port")
var verb = flag.Int("v", 2, "log verbosity")

var (
	// VERSION is injected by buildflags
	VERSION = "SELFBUILD"
)

//type OnLoadFunc func( plugin_name, plugin_base_path string ) (hand http.Handler, err error)
//type OnStopFunc func( cleanup bool ) (ok bool, err error)

type plugin struct {
	PluginName  *string
	Version     *string
//	loadFunc    OnLoadFunc
//	stopFunc    OnStopFunc
//	loadFunc    func( string, string ) ( http.Handler, error )
	loadFunc    func( string, string, web.Session ) ( http.Handler, error )
	stopFunc    func( bool ) ( bool, error )
	mx          sync.Mutex
}

func main() {
	if VERSION == "SELFBUILD" {
		std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}
	flag.Parse()

	Verbosity = *verb
	Vln(2, "version:", VERSION)

	lis, err := net.Listen("tcp", *bind)
	if err != nil {
		Vln(2, "Error listening:", err.Error())
		os.Exit(1)
	}
	defer lis.Close()

	Vln(2, "listening on:", lis.Addr())
	Vln(2, "verbosity:", *verb)

	sess, _ := web.NewManager()
	sess.Maxlifetime = 30 * time.Second
	sess.Gclifetime = 10 * time.Second
	mainsess, _ := sess.NewPlugin("/")
	go sess.SessionGC()
	defer sess.StopGC()

	mux := http.NewServeMux()

	rh := http.RedirectHandler("http://example.org", 307)
	mux.Handle("/foo", rh)

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		store, err := mainsess.Start(w, r)
		if err != nil {
			fmt.Fprintf(w, "SessionStart error! <br/>\n")
		}
		fmt.Fprintf(w, "Welcome to the home page! %s\n", r.URL.Path)
		fmt.Fprintf(w, "Session isLogin = %v\n", store.IsLogin())
		store.Set("isLogin", true)
	})

	loaded := make([]*plugin, 0)
	plugins, _ := ioutil.ReadDir("plugins/")
	for _, p := range plugins {

		pname := p.Name()
		pbase := path.Join("plugins/", pname)
		Vln(2, "[load-plugin]", pname, pbase)

		plug, err := loadPlugin(pname, pbase)
		if err != nil {
			Vln(2, "[load-plugin][Fail]", pname, err)
			continue
		}

		plugsess, _ := sess.NewPlugin(*plug.PluginName + "-v" + *plug.Version + "@" + pname)
		hand, err := plug.loadFunc(pname, pbase, plugsess)
		if err != nil {
			Vln(2, "[load-plugin][Load]err", pname, err)
			continue
		}
		mux.Handle("/" + pname + "/", hand)

		Vln(2, "[load-plugin][OK]", pname)

		loaded = append(loaded, plug)
	}

	http.Serve(lis, mux)
}

func loadPlugin(pname, pbase string) (*plugin, error) {
	errType := errors.New("wrong type")

	lib, err := PL.Open(path.Join(pbase, "plugin.so"))
	if err != nil {
		Vln(2, "[load-plugin]open err", pname, err)
		return nil, err
	}

	name, err := lib.Lookup("PLIGIN_NAME")
	if err != nil {
		Vln(2, "[load-plugin][PLIGIN_NAME]Lookup err", pname, err)
		return nil, err
	}
	plugname, ok := name.(*string)
	if !ok {
		Vln(2, "[load-plugin][PLIGIN_NAME]wrong type", pname)
		return nil, errType
	}

	ver, err := lib.Lookup("VERSION")
	if err != nil {
		Vln(2, "[load-plugin][VERSION]Lookup err", pname, err)
		return nil, err
	}
	version, ok := ver.(*string)
	if !ok {
		Vln(2, "[load-plugin][VERSION]wrong type", pname)
		return nil, errType
	}

	load, err := lib.Lookup("OnLoad")
	if err != nil {
		Vln(2, "[load-plugin][OnLoad]Lookup err", pname, err)
		return nil, err
	}
	loadImpl, ok := load.(func( string, string, web.Session ) ( http.Handler, error ))
	if !ok {
		Vln(2, "[load-plugin][OnLoad]wrong type", pname)
		return nil, errType
	}

	stop, err := lib.Lookup("OnStop")
	if err != nil {
		Vln(2, "[load-plugin][OnStop]Lookup err", pname, err)
		return nil, err
	}
	stopImpl, ok := stop.(func( bool ) ( bool, error ))
	if !ok {
		Vln(2, "[load-plugin][OnStop]wrong type", pname)
		return nil, errType
	}

	plug := &plugin {
		PluginName: plugname,
		Version: version,
		loadFunc: loadImpl,
		stopFunc: stopImpl,
	}

	return plug, nil
}

var Verbosity int = 2
var std = log.New(os.Stderr, "", log.LstdFlags)

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

