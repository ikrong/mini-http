package src

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

var (
	httpPort  = flag.Int("port", 80, "HTTP Port")
	httpsPort = flag.Int("https-port", 0, "HTTPS Port")
	wwwRoot   = flag.String("root", "/www/", "WWW Root")
	_         = flag.String("domain", "", "Domain")
	_         = flag.String("cert", "", "Domain Cert File")
	_         = flag.String("key", "", "Domain Key File")
	_         = flag.String("mode", "", "Can Be Set To 'history' to Support Web APP Routing")
	_         = flag.String("proxy", "", "Set proxy api")
	_         = flag.Bool("skip-tls-verify", true, "Skip tls verify")
	notFound  = flag.String("not-found", "/404.html", "Custom 404 page")
	domains   []DomainConfig
)

func RunServer() {
	if !flag.Parsed() {
		flag.Parse()
	}

	domains = ParseDomains(*wwwRoot)

	fmt.Println("Mini HTTP is Starting")

	handler := &StaticServerHandler{
		Domains:        domains,
		DefaultWWWRoot: *wwwRoot,
		NotFound:       *notFound,
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: handler,
	}

	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	if *httpsPort > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *httpsPort))
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig := &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				domain := CurrentDomain(&domains, chi.ServerName)
				if domain == nil {
					return nil, errors.New("no certificate found")
				}
				if domain.Cert == "" {
					err := getRootCertificate()
					if err != nil {
						fmt.Println("root certificate generate failed", err)
						return nil, err
					}
					cert, err := getDomainCertificate(domain.Domain)
					if err != nil {
						fmt.Println("domain certificate generate failed", err)
						return nil, err
					}
					return cert, nil
				}
				cert, err := tls.LoadX509KeyPair(domain.Cert, domain.Key)
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
		}
		go func() {
			log.Fatal(http.Serve(&TLSServerListener{
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
			})))
		}()
	}

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	fmt.Println("Try Pressing CTRL + C to Shutdown")
	select {
	case <-sigChannel:
		fmt.Println("")
		fmt.Println("Mini HTTP Closed")
		os.Exit(0)
	}
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
