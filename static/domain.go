package static

import (
	"crypto/tls"
	"fmt"
	"net/http/httputil"
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

func NewDomain() (domain DomainConfig) {
	domain = DomainConfig{
		Root:     "/www",
		NotFound: "/404.html",
	}
	return
}

func (d *DomainConfig) label() (label string) {
	label = "default"
	if d.Domain != "" {
		label = d.Domain
	}
	return
}

func (d *DomainConfig) isEmpty() (empty bool) {
	empty = true
	if d.Domain != "" {
		empty = false
	}
	return
}

func (d *DomainConfig) print() {
	fmt.Printf("%s: \t%s\n", d.label(), d.Root)
	fmt.Printf("\t404: \t%s\n", d.NotFound)
	if d.Mode != "" {
		fmt.Printf("\tMode: \t%s\n", d.Mode)
	}
	if d.Cert != "" {
		fmt.Printf("\tCert: \t%s\n", d.Cert)
	}
	if d.Key != "" {
		fmt.Printf("\tKey: \t%s\n", d.Key)
	}
	if d.Proxy != nil {
		for _, proxy := range *d.Proxy {
			fmt.Printf("\tProxy: \t%s -> %s\n", proxy.Url, proxy.Proxy)
		}
	}
}

func (d *DomainConfig) loadCertificate() (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(d.Cert, d.Key)
	if err != nil {
		return nil, err
	}
	return &cert, err
}
