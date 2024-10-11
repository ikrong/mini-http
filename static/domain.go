package static

import (
	"crypto/tls"
	"fmt"
	"mini-http/log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type DomainProxy struct {
	Url      string
	Proxy    string
	Instance *httputil.ReverseProxy
}

type DomainConfig struct {
	Domain       string
	Cert         string
	Key          string
	Mode         string
	Root         string
	NotFound     string
	Proxy        *[]DomainProxy
	AutoProxy    bool
	AutoProxyKey *string
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
	if d.AutoProxy {
		fmt.Println("\tAutoProxy: \tEnabled")
	}
	if d.AutoProxyKey != nil && *d.AutoProxyKey != "" {
		fmt.Printf("\tAutoProxyKey: \t%s\n", *d.AutoProxyKey)
	}
	if d.Proxy != nil {
		for _, proxy := range *d.Proxy {
			fmt.Printf("\tProxy: \t%s --> %s\n", proxy.Url, proxy.Proxy)
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

func (s *DomainConfig) readAutoProxyConfig(r *http.Request) {
	if !s.AutoProxy {
		return
	}
	key := "proxyconfig"
	if s.AutoProxyKey != nil && *s.AutoProxyKey != "" {
		key = *s.AutoProxyKey
	}
	rawCookie, err := r.Cookie(key)
	if err != nil {
		return
	}
	cookie, _ := url.QueryUnescape(rawCookie.Value)
	if cookie == "" {
		return
	}
	proxyList := make([]DomainProxy, 0)
	changed := false
	for i, c := range strings.Split(cookie, ";") {
		pUrl, pProxy, exist := strings.Cut(strings.TrimSpace(c), ":")
		if !exist {
			continue
		}
		p := DomainProxy{
			Url:   pUrl,
			Proxy: pProxy,
		}
		proxyList = append(proxyList, p)
		if !changed {
			if s.Proxy != nil && i < len(*s.Proxy) && (*s.Proxy)[i].Url == p.Url && (*s.Proxy)[i].Proxy == p.Proxy {
				continue
			} else {
				changed = true
			}
		}
	}
	if changed {
		s.Proxy = &proxyList
		log.Info("AutoProxy Configuration Changed")
	}
}
