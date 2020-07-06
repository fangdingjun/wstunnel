package main

type proxyItem struct {
	Listen string `yaml:"listen"`
	Remote string `yaml:"remote"`
}

type conf struct {
	ProxyConfig []proxyItem `yaml:"proxy_config"`
}
