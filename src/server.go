package src

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

func RunServer(args []string) (err error) {
	serverConfig := ServerConfig{
		HTTPPort:      80,
		HTTPSPort:     0,
		Domains:       []DomainConfig{},
		DefaultDomain: NewDomain(),
	}
	serverConfig.ParseFromArgs(args)

	fmt.Println("Starting Mini HTTP...")

	handler := &StaticServerHandler{
		serverConfig: serverConfig,
	}

	fmt.Printf("Listen TCP: ")
	if serverConfig.HTTPPort > 0 {
		fmt.Printf("%d ", serverConfig.HTTPPort)
	}
	if serverConfig.HTTPSPort > 0 {
		fmt.Printf("%d ", serverConfig.HTTPSPort)
	}
	fmt.Println("")
	serverConfig.PrintConfig()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", serverConfig.HTTPPort))
	if err != nil {
		log.Panic(err)
		return
	}

	go func() {
		if err := http.Serve(ln, handler); err != nil {
			log.Panic(err)
		}
	}()

	if serverConfig.HTTPSPort > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", serverConfig.HTTPSPort))
		if err != nil {
			log.Panic(err)
		}
		tlsConfig := &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				domain := serverConfig.CurrentDomain(chi.ServerName)
				return domain.loadCertificate()
			},
		}
		go func() {
			if err = http.Serve(&TLSServerListener{
				Listener:  ln,
				TlsConfig: tlsConfig,
			}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.TLS == nil {
					u := url.URL{
						Scheme:   "https",
						Opaque:   r.URL.Opaque,
						User:     r.URL.User,
						Host:     r.Host,
						Path:     r.URL.Path,
						RawQuery: r.URL.RawQuery,
						Fragment: r.URL.Fragment,
					}
					// 如果通过http访问，则自动重定向到https
					http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
				} else {
					handler.ServeHTTP(w, r)
				}
			})); err != nil {
				log.Panic(err)
			}
		}()
	}

	return
}

type Conn struct {
	net.Conn
	b byte
	e error
	f bool
}

func (c *Conn) Read(b []byte) (int, error) {
	if c.f {
		c.f = false
		b[0] = c.b
		if len(b) > 1 && c.e == nil {
			n, e := c.Conn.Read(b[1:])
			if e != nil {
				c.Conn.Close()
			}
			return n + 1, e
		} else {
			return 1, c.e
		}
	}
	return c.Conn.Read(b)
}

type TLSServerListener struct {
	TlsConfig *tls.Config
	net.Listener
}

func (l *TLSServerListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	b := make([]byte, 1)
	_, err = c.Read(b)
	if err != nil {
		c.Close()
		if err != io.EOF {
			return nil, err
		}
	}

	con := &Conn{
		Conn: c,
		b:    b[0],
		e:    err,
		f:    true,
	}

	if b[0] == 22 {
		// 如果请求是https，则开始使用证书握手
		return tls.Server(con, l.TlsConfig), nil
	}

	// 否则是http请求
	return con, nil
}
