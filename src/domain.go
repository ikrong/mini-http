package src

import (
	"net/http/httputil"
	"os"
	"strings"
)

type DomainProxy struct {
	Url      string
	Proxy    string
	Instance *httputil.ReverseProxy
}

type DomainConfig struct {
	Domain   string
	Cert     string
	Key      string
	Mode     string
	Root     string
	NotFound string
	Proxy    *[]DomainProxy
}

func ParseDomains(wwwRoot string) (domains []DomainConfig) {
	var domain DomainConfig
	for i := 0; i < len(os.Args); i++ {
		if i+1 <= len(os.Args) {
			var key = os.Args[i]
			switch {
			case key == "--domain":
				if domain != (DomainConfig{}) && domain.Domain != "" {
					domains = append(domains, domain)
				}
				domain = DomainConfig{Domain: os.Args[i+1], Root: wwwRoot}
				i += 1
			case key == "--cert":
				domain.Cert = os.Args[i+1]
				i += 1
			case key == "--key":
				domain.Key = os.Args[i+1]
				i += 1
			case key == "--mode":
				domain.Mode = os.Args[i+1]
				i += 1
			case key == "--root":
				domain.Root = os.Args[i+1]
				i += 1
			case key == "--proxy":
				if domain.Proxy == nil {
					p := (make([]DomainProxy, 0))
					domain.Proxy = &p
				}
				proxy := append(*domain.Proxy, parseDomainProxy(os.Args[i+1]))
				domain.Proxy = &proxy
				i += 1
			case key == "--not-found":
				domain.NotFound = os.Args[i+1]
				i += 1
			}
		}
	}
	if domain != (DomainConfig{}) && domain.Domain != "" {
		domains = append(domains, domain)
	}
	return
}

func parseDomainProxy(cmd string) DomainProxy {
	index := strings.Index(cmd, ":")
	return DomainProxy{Url: cmd[0:index], Proxy: cmd[index+1:]}
}

func CurrentDomain(domains *[]DomainConfig, host string) (domain *DomainConfig) {
	hostInfos := strings.Split(host, ":")
	if len(*domains) > 0 {
		domain = &(*domains)[0]
	}
	for i := 0; i < len(*domains); i++ {
		if (*domains)[i].Domain == hostInfos[0] {
			domain = &(*domains)[i]
			return
		}
	}
	return domain
}
