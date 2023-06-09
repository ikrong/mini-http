package src

import (
	"os"
	"strings"
)

type DomainConfig struct {
	Domain   string
	Cert     string
	Key      string
	Mode     string
	Root     string
	NotFound string
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

func CurrentDomain(domains *[]DomainConfig, host string) (domain *DomainConfig) {
	hostInfos := strings.Split(host, ":")
	for i := 0; i < len(*domains); i++ {
		if (*domains)[i].Domain == hostInfos[0] {
			domain = &(*domains)[i]
			return
		}
	}
	return nil
}
