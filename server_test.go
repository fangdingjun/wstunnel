package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	log "github.com/fangdingjun/go-log/v5"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
)

func TestServer(t *testing.T) {
	cfgfile := "config.example.yaml"
	data, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	var cfg conf
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}
	makeServers(cfg)
	l1, err := net.Listen("tcp", "127.0.0.1:2903")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer l1.Close()
		for {

			c1, err := l1.Accept()
			if err != nil {
				log.Errorln(err)
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				data := make([]byte, 1024)
				for {
					n, err := c.Read(data)
					if err != nil {
						log.Errorln(err)
						break
					}
					c.Write(data[:n])
					log.Infof("2903 receive: %s", string(data[:n]))
				}
			}(c1)
		}
	}()
	l2, err := net.Listen("tcp", "127.0.0.1:2904")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer l2.Close()
		for {
			c1, err := l2.Accept()
			if err != nil {
				log.Errorln(err)
			}
			go func(c net.Conn) {
				defer c.Close()
				data := make([]byte, 1024)
				for {
					n, err := c.Read(data)
					if err != nil {
						log.Errorln(err)
						break
					}
					c.Write(data[:n])
					log.Infof("2904 receive: %s", string(data[:n]))
				}
			}(c1)
		}
	}()

	time.Sleep(time.Second)
	c1, resp, err := dialer.Dial("ws://127.0.0.1:2901/p1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("dial ws code %d", resp.StatusCode)
	}
	err = c1.WriteMessage(websocket.BinaryMessage, []byte("p1"))
	if err != nil {
		t.Fatal(err)
	}
	_, d, err := c1.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte("p1"), d) {
		t.Errorf("failed msg not equal, expect p1, got %s", d)
	}
	c2, resp, err := dialer.Dial("ws://127.0.0.1:2901/p2", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("dial ws code %d", resp.StatusCode)
	}
	err = c2.WriteMessage(websocket.BinaryMessage, []byte("p2"))
	if err != nil {
		t.Fatal(err)
	}
	_, d, err = c2.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte("p2"), d) {
		t.Errorf("failed msg not equal, expect p2, got %s", d)
	}

	c3, err := net.Dial("tcp", "127.0.0.1:2905")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c3.Write([]byte("c3"))
	if err != nil {
		t.Fatal(err)
	}
	d2 := make([]byte, 20)
	n, err := c3.Read(d2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte("c3"), d2[:n]) {
		t.Errorf("failed msg not equal, expect c3, got %s", d2[:n])
	}
}
