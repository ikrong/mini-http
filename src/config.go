package src

import (
	"fmt"
	"strconv"
	"strings"
)

type ServerConfig struct {
	HTTPPort      int
	HTTPSPort     int
	Domains       []DomainConfig
	DefaultDomain DomainConfig
}

func (c *ServerConfig) ParseFromArgs(args []string) {
	var domain = NewDomain()
	for i := 0; i < len(args); i++ {
		if i+1 <= len(args) {
			var key = args[i]
			switch {
			case key == "--domain":
				if !domain.isEmpty() {
					c.Domains = append(c.Domains, domain)
				} else {
					c.DefaultDomain = domain
				}
				domain = NewDomain()
				domain.Domain = args[i+1]
				i += 1
			case key == "--cert":
				domain.Cert = args[i+1]
				i += 1
			case key == "--key":
				domain.Key = args[i+1]
				i += 1
			case key == "--mode":
				if args[i+1] == "history" {
					domain.Mode = args[i+1]
				}
				i += 1
			case key == "--root":
				domain.Root = args[i+1]
				i += 1
			case key == "--proxy":
				if domain.Proxy == nil {
					p := (make([]DomainProxy, 0))
					domain.Proxy = &p
				}
				proxy := append(*domain.Proxy, c.parseDomainProxy(args[i+1]))
				domain.Proxy = &proxy
				i += 1
			case key == "--not-found":
				domain.NotFound = args[i+1]
				i += 1
			case key == "--port":
				port, _ := strconv.ParseInt(args[i+1], 0, strconv.IntSize)
				c.HTTPPort = int(port)
				i += 1
			case key == "--https-port":
				port, _ := strconv.ParseInt(args[i+1], 0, strconv.IntSize)
				c.HTTPSPort = int(port)
				i += 1
			}
		}
	}
	if !domain.isEmpty() {
		c.Domains = append(c.Domains, domain)
	} else {
		c.DefaultDomain = domain
	}
}

func (c *ServerConfig) parseDomainProxy(cmd string) DomainProxy {
	index := strings.Index(cmd, ":")
	return DomainProxy{Url: cmd[0:index], Proxy: cmd[index+1:]}
}

func (c *ServerConfig) PrintConfig() {
	// 将所有domains以表格形式输出到控制台
	fmt.Println("Static Server Configuration:")
	c.DefaultDomain.print()
	for _, domain := range c.Domains {
		domain.print()
	}
	fmt.Println("")
}

func (s *ServerConfig) CurrentDomain(host string) (domain DomainConfig) {
	hostInfos := strings.Split(host, ":")
	domain = s.DefaultDomain
	if len(s.Domains) > 0 {
		domain = s.Domains[0]
	}
	for i := 0; i < len(s.Domains); i++ {
		if (s.Domains)[i].Domain == hostInfos[0] {
			domain = s.Domains[i]
			return
		}
	}
	return domain
}
