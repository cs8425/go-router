# go-router
a base system design for Soft-routing base on go-plugin which provide a simple web interface to use.
Proposal are minimum dependencies, minimum permission, easy install, as safe as it can.

為軟路由設計的路由器UI系統, 以安全、易安裝、最低權限、最小相依為目標, 基於golang plugin.

## main frame
* permission
	* login
	* chpwd
	* logout
* self log, self status
	* runtime.NumCPU()
	* runtime.NumGoroutine()
	* runtime.NumCgoCall()
	* runtime.ReadMemStats(m *MemStats)
* power control
	* reboot
	* poweroff
* plugin manager

## core plugin/function (only std package)
* PPPoE, DCHP, IP setting, DHCP server, route
	* dhcpclient
	* [dnsmasq](https://wiki.debian.org/HowTo/dnsmasq)
	* PPPoE
		* /etc/ppp/peers/dsl-provider
		* plog => 查看 PPPoE 連線最近的 log，故障排除時很有用
		* poff => 斷線
		* pon dsl-provider => 啟動上述設定的連線
* iptables
	* `sudo` with `iptables` permission
* system status
	* by parsing /proc/*
* consloe
	* any shell

### Requirements
* a tcp port to bind
* a user and group
* `sudo`, for network setting (`ip`, `iptables`), power control (shutdown, reboot)
* `can bind < 1024 port`
* OpenVZ needs vSwap


### Directory
* /config.json
* /res/
* /templ/
* /plugins/
	* /[name]
		* res/
		* templ/
		* plugin.so
		* info.json : ro, long name, version, info
		* config.json : rw

### plugin methods

* session
	* [x] IsLogin() bool
	* [ ] HasPermission() bool
	* [ ] GetCSRF() string

* plugin.so
	* var PLIGIN_NAME string
	* var VERSION string
	* func onLoad( plugin_name, plugin_base_path string, session web.Session, mux *web.Mux ) (hand http.Handler, err error)
		* initial plugin and return `http.Handler` for web UI.
	* func OnStop( cleanup bool ) (ok bool, err error)
		* system exit (rollback or not depend on `cleanup`), clean up all goroutine.


### plugin to implement
* port knocking
* fail2ban
* let's encrypt
* DDNS
* cron, at
* system/... log viewer
* QoS
* system services manager
* browser via image stream by phantomjs backend
* noVNC
* docker manager
* BT

## known issue
* plugin.so big file size (x86_64, ~10MB) (x86_64, `-ldflags=-s -w`, ~6MB)
* building plugin very slow
* load lots of plugin cause huge ram usage

