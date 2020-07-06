package main

import (
	"io"
	"net"
	"net/url"

	log "github.com/fangdingjun/go-log/v5"
	"github.com/gorilla/websocket"
)

func forwardWS2WS(conn1, conn2 *websocket.Conn) {
	ch := make(chan struct{}, 2)

	go func() {
		for {
			t, data, err := conn1.ReadMessage()
			if err != nil {
				log.Errorln(err)
				break
			}
			err = conn2.WriteMessage(t, data)
			if err != nil {
				log.Errorln(err)
				break
			}
		}
		ch <- struct{}{}
	}()

	go func() {
		for {
			t, data, err := conn2.ReadMessage()
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

func makeServers(cfg conf) {
	var wsservers = []wsServer{}
	var tcpservers = []tcpServer{}

	for _, c := range cfg.ProxyConfig {
		u, err := url.Parse(c.Listen)
		if err != nil {
			log.Fatalf("parse %s, error %s", c.Listen, err)
		}

		switch u.Scheme {
		case "ws":
			exists := false
			for i := 0; i < len(wsservers); i++ {
				if wsservers[i].addr == u.Host {
					exists = true
					wsservers[i].rule = append(wsservers[i].rule, forwardRule{u.Path, c.Remote})
					break
				}
			}
			if !exists {
				wsservers = append(wsservers, wsServer{u.Host, []forwardRule{{u.Path, c.Remote}}})
			}
		case "tcp":
			tcpservers = append(tcpservers, tcpServer{u.Host, c.Remote})
		default:
			log.Fatalf("unsupported scheme %s", u.Scheme)
		}
	}

	for _, srv := range wsservers {
		go srv.run()
	}

	for _, srv := range tcpservers {
		go srv.run()
	}
}
