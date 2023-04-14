package src

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	httpPort  = flag.Int("port", 80, "HTTP Port")
	httpsPort = flag.Int("https-port", 443, "HTTPS Port")
	wwwRoot   = flag.String("root", "/www/", "WWW Root")
	_         = flag.String("domain", "", "Domain")
	_         = flag.String("cert", "", "Domain Cert File")
	_         = flag.String("key", "", "Domain Key File")
	_         = flag.String("mode", "", "Can Be Set To 'history' to Support Web APP Routing")
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
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: handler,
	}

	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	if *httpsPort > 0 {
		httpsServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", *httpsPort),
			Handler: handler,
			TLSConfig: &tls.Config{
				GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
					domain := CurrentDomain(&domains, chi.ServerName)
					if domain == nil {
						return nil, errors.New("no certificate found")
					}
					cert, err := tls.LoadX509KeyPair(domain.Cert, domain.Key)
					if err != nil {
						return nil, err
					}
					return &cert, nil
				},
			},
		}

		go func() {
			log.Fatal(httpsServer.ListenAndServeTLS("", ""))
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
