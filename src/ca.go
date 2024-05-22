package src

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"sync"
	"time"
)

var (
	rootCertPEM  []byte
	rootKeyPEM   []byte
	certCache    sync.Map
	rootCertPath = "/certs/root.crt"
	rootKeyPath  = "/certs/root.key"
)

func getRootCertificate() (err error) {
	if _, err = os.Stat(rootCertPath); err == nil {
		certPEM, _ := os.ReadFile(rootCertPath)
		keyPEM, _ := os.ReadFile(rootKeyPath)
		rootCertPEM = certPEM
		rootKeyPEM = keyPEM
		return
	}
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	notBefore := time.Now()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"IKrong Root CA"}},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pemBlock, _ := x509.MarshalECPrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: pemBlock})

	err = os.MkdirAll("/certs", 0755)
	if err != nil {
		return
	}
	err = os.WriteFile(rootCertPath, certPEM, 0644)
	if err != nil {
		return
	}
	err = os.WriteFile(rootKeyPath, keyPEM, 0644)
	if err != nil {
		return
	}
	rootCertPEM = certPEM
	rootKeyPEM = keyPEM
	return
}

func getDomainCertificate(domain string) (*tls.Certificate, error) {
	if cert, ok := certCache.Load(domain); ok {
		return cert.(*tls.Certificate), nil
	}

	cert, err := generateDomainCertificate(domain)
	if err != nil {
		return nil, err
	}

	certCache.Store(domain, cert)
	return cert, nil
}

func generateDomainCertificate(domain string) (*tls.Certificate, error) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{domain}},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	rootCertBlock, _ := pem.Decode(rootCertPEM)
	rootCert, _ := x509.ParseCertificate(rootCertBlock.Bytes)

	rootKeyBlock, _ := pem.Decode(rootKeyPEM)
	rootKey, _ := x509.ParseECPrivateKey(rootKeyBlock.Bytes)

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, rootCert, &priv.PublicKey, rootKey)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	pemBlock, _ := x509.MarshalECPrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: pemBlock})

	cert, _ := tls.X509KeyPair(certPEM, keyPEM)

	return &cert, nil
}
