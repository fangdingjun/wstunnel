package main

import (
	"io"
	"net"
	"net/http"
	"net/url"

	log "github.com/fangdingjun/go-log/v5"
	"github.com/gorilla/websocket"
)

type forwardRule struct {
	local  string
	remote string
}

type wsServer struct {
	addr string
	rule []forwardRule
}

type tcpServer struct {
	addr   string
	remote string
}

func (wss *wsServer) run() {
	if err := http.ListenAndServe(wss.addr, wss); err != nil {
		log.Errorln(err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var dialer = &websocket.Dialer{}

func forwardWS2WS(conn, conn1 *websocket.Conn) {
	ch := make(chan struct{}, 2)

	go func() {
		for {
			t, data, err := conn.ReadMessage()
			if err != nil {
				log.Errorln(err)
				break
			}
			err = conn1.WriteMessage(t, data)
			if err != nil {
				log.Errorln(err)
				break
			}
		}
		ch <- struct{}{}
	}()

	go func() {
		for {
			t, data, err := conn1.ReadMessage()
			if err != nil {
				log.Errorln(err)
				break
			}
			err = conn.WriteMessage(t, data)
			if err != nil {
				log.Errorln(err)
				break
			}
		}
		ch <- struct{}{}
	}()

	<-ch
}

func forwardWS2TCP(conn1 *websocket.Conn, conn2 net.Conn) {
	ch := make(chan struct{}, 2)

	go func() {
		for {
			_, data, err := conn1.ReadMessage()
			if err != nil {
				log.Errorln(err)
				break
			}

			_, err = conn2.Write(data)
			if err != nil {
				log.Errorln(err)
				break
			}
		}
		ch <- struct{}{}
	}()

	go func() {
		buf := make([]byte, 1024)

		for {
			n, err := conn2.Read(buf)
			if err != nil {
				log.Errorln(err)
				break
			}

			err = conn1.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				log.Errorln(err)
				break
			}
		}
		ch <- struct{}{}
	}()

	<-ch
}

func forwardTCP2TCP(c1, c2 net.Conn) {
	ch := make(chan struct{}, 2)

	go func() {
		_, err := io.Copy(c1, c2)
		if err != nil {
			log.Errorln(err)
		}
		ch <- struct{}{}
	}()

	go func() {
		_, err := io.Copy(c2, c1)
		if err != nil {
			log.Errorln(err)
		}
		ch <- struct{}{}
	}()

	<-ch
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
