package main

import (
	"net"
	"net/http"
	"net/url"

	log "github.com/fangdingjun/go-log/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var dialer = &websocket.Dialer{}

type forwardRule struct {
	local  string
	remote string
}

type wsServer struct {
	addr string
	rule []forwardRule
}

func (wss *wsServer) run() {
	if err := http.ListenAndServe(wss.addr, wss); err != nil {
		log.Errorln(err)
	}
}

func (wss *wsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	remote := ""
	for _, ru := range wss.rule {
		if ru.local == p {
			remote = ru.remote
		}
	}

	if remote == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	ip := r.RemoteAddr

	_ip := r.Header.Get("x-real-ip")
	if _ip != "" {
		ip = _ip
	}

	log.Debugf("from %s, request %s, forward to %s", ip, p, remote)

	defer func() {
		log.Debugf("from %s, request finished", ip)
	}()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	u, _ := url.Parse(remote)

	if u.Scheme == "ws" || u.Scheme == "wss" {
		conn1, resp, err := dialer.Dial(remote, nil)
		if err != nil {
			log.Errorln(err)
			return
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			log.Errorf("dial remote ws %d", resp.StatusCode)
			return
		}
		defer conn1.Close()

		forwardWS2WS(conn, conn1)
		return
	}

	if u.Scheme == "tcp" {
		conn1, err := net.Dial("tcp", u.Host)
		if err != nil {
			log.Errorln(err)
			return
		}
		defer conn1.Close()

		forwardWS2TCP(conn, conn1)
		return
	}
	log.Errorf("unsupported scheme %s", u.Scheme)
}
