package main

import (
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	log "github.com/fangdingjun/go-log/v5"
	"gopkg.in/yaml.v2"
)

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

func main() {
	var cfgfile string
	var logfile string
	var loglevel string
	flag.StringVar(&cfgfile, "c", "config.yaml", "config file")
	flag.StringVar(&logfile, "log_file", "", "log file")
	flag.StringVar(&loglevel, "log_level", "INFO", "log level")
	flag.Parse()

	data, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	var cfg conf
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	if logfile != "" {
		log.Default.Out = &log.FixedSizeFileWriter{
			MaxCount: 4,
			Name:     logfile,
			MaxSize:  10 * 1024 * 1024,
		}
	}

	if lv, err := log.ParseLevel(loglevel); err == nil {
		log.Default.Level = lv
	}

	makeServers(cfg)

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-ch:
		log.Printf("received signal %s, exit.", s)
	}
}
