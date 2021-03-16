package main

import (
	"net"
	"net/http"
	"net/url"

	log "github.com/fangdingjun/go-log/v5"
)

type tcpServer struct {
	addr   string
	remote string
}

func (srv *tcpServer) run() {
	l, err := net.Listen("tcp", srv.addr)
	if err != nil {
		log.Errorln(err)
		return
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go srv.serve(conn)
	}
}

func (srv *tcpServer) serve(c net.Conn) {
	defer c.Close()

	u, _ := url.Parse(srv.remote)

	log.Debugf("connected from %s, forward to %s", c.RemoteAddr(), srv.remote)

	defer func() {
		log.Debugf("from %s, finished", c.RemoteAddr())
	}()

	if u.Scheme == "ws" || u.Scheme == "wss" {
		conn1, resp, err := dialer.Dial(srv.remote, nil)
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

		forwardWS2TCP(conn1, c)
		return
	}

	if u.Scheme == "tcp" {
		conn1, err := net.Dial("tcp", u.Host)
		if err != nil {
			log.Errorln(err)
			return
		}
		defer conn1.Close()

		forwardTCP2TCP(c, conn1)
		return
	}

	log.Errorf("unsupported scheme %s", u.Scheme)
}
