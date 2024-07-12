package static

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
	"path"
	"sync"
	"time"
)

var ca = CA{}

type CA struct {
	certByte []byte
	keyByte  []byte
	store    sync.Map
	mu       sync.Mutex
}

func (ca *CA) getRootDir() (rootDir string, err error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		userDir = "/tmp/"
	}
	rootDir = path.Join(userDir, "ikrong/mini-http")
	if _, err = os.Stat(rootDir); err != nil {
		os.MkdirAll(rootDir, 0755)
		err = nil
	}
	return
}

func (ca *CA) generateRootCertificate() (err error) {
	if ca.certByte != nil {
		return
	}
	rootDir, err := ca.getRootDir()
	if err != nil {
		return
	}
	certFile := path.Join(rootDir, "root.crt")
	keyFile := path.Join(rootDir, "root.key")
	var cert []byte
	var key []byte
	if _, err = os.Stat(certFile); err == nil {
		if cert, err = os.ReadFile(certFile); err == nil {
			if key, err = os.ReadFile(keyFile); err == nil {
				ca.certByte = cert
				ca.keyByte = key
				return
			}
		}
	}

	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	notBefore := time.Now()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	// 设置证书模板
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"IKrong Root CA"},
			CommonName:   "IKrong Root CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pemBlock, _ := x509.MarshalECPrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: pemBlock})

	if err = os.WriteFile(certFile, certPEM, 0644); err != nil {
		return
	}

	if err = os.WriteFile(keyFile, keyPEM, 0644); err != nil {
		return
	}
	ca.certByte = certPEM
	ca.keyByte = keyPEM
	return
}

func (ca *CA) issueCertificate(domain string) (*tls.Certificate, error) {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	if err := ca.generateRootCertificate(); err != nil {
		return nil, err
	}
	if cert, ok := ca.store.Load(domain); ok {
		return cert.(*tls.Certificate), nil
	}

	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	rootCertBlock, _ := pem.Decode(ca.certByte)
	rootCert, _ := x509.ParseCertificate(rootCertBlock.Bytes)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{domain},
			CommonName:   domain,
		},
		Issuer:                rootCert.Subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	rootKeyBlock, _ := pem.Decode(ca.keyByte)
	rootKey, _ := x509.ParseECPrivateKey(rootKeyBlock.Bytes)

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, rootCert, &priv.PublicKey, rootKey)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	pemBlock, _ := x509.MarshalECPrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: pemBlock})

	cert, _ := tls.X509KeyPair(certPEM, keyPEM)

	ca.store.Store(domain, &cert)

	return &cert, nil
}
