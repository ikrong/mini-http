package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"mini-http/static"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var port = 32000
var currentDir string

type testRequest struct {
	url      string
	response string
	status   int
	allowErr bool
}

type testCase struct {
	label               string
	enableSelfSignedSSL bool
	ca                  string
	cert                string
	key                 string
	args                []string
	requests            []testRequest
}

func (c *testCase) Run(t *testing.T) {
	fmt.Println("================== Start Test " + c.label + " ==================")
	t.Run(c.label, func(t *testing.T) {
		var ports []string
		var httpPort int
		var httpsPort int
		port++
		httpPort = port
		ports = append(ports, "--port", fmt.Sprintf("%d", httpPort))

		if c.cert != "" && c.key != "" {
			port++
			httpsPort = port
			ports = append(ports, "--https-port", fmt.Sprintf("%d", httpsPort), "--cert", c.cert, "--key", c.key)
		} else if c.enableSelfSignedSSL {
			port++
			httpsPort = port
			ports = append(ports, "--https-port", fmt.Sprintf("%d", httpsPort))
		}

		args := append(ports, c.args...)

		err := static.RunServer(args)
		if err != nil {
			t.Fatal(err)
		}

		for _, request := range c.requests {
			url := fmt.Sprintf(request.url, httpPort)
			if strings.Contains(request.url, "https") {
				url = fmt.Sprintf(request.url, httpsPort)
			}
			content, status, err := get(url, c.ca)
			if err != nil {
				if request.allowErr {
					continue
				}
				t.Error(err)
				continue
			}
			if request.response != "" {
				assert.Equal(t, request.response, content)
			}
			if request.status != 0 {
				assert.Equal(t, request.status, status)
			}
		}
	})
	fmt.Println("================== End Test " + c.label + " ==================")
}

func getSelfSignedCertDir() string {
	userDir, err := os.UserConfigDir()
	if err != nil {
		userDir = "/tmp/"
	}
	dir := path.Join(userDir, "ikrong/mini-http")
	return dir
}

func get(url string, cert string) (content string, status int, err error) {
	transport := &http.Transport{}
	if strings.Contains(url, "https") {
		if cert == "" {
			cert = path.Join(getSelfSignedCertDir(), "root.crt")
		}
		var certContent []byte
		certContent, err = os.ReadFile(cert)
		if err == nil {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(certContent)
			transport.TLSClientConfig = &tls.Config{RootCAs: caCertPool}
		}
	}

	client := &http.Client{Transport: transport}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	status = response.StatusCode
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	content = string(body)
	return
}

func Test(t *testing.T) {
	var testCaseList = []testCase{
		{
			label: "Test Default Server",
			args: []string{
				"--root", fmt.Sprintf("%s/assets/", currentDir),
				"--not-found", fmt.Sprintf("%s/assets/domain/example.net/index.html", currentDir),
			},
			requests: []testRequest{
				{
					url:    "http://localhost:%d",
					status: http.StatusOK,
				},
				{
					url:      "http://localhost:%d/404/",
					status:   http.StatusNotFound,
					response: "example.net",
				},
			},
		},
		{
			label: "Test Domain Server",
			args: []string{
				"--domain", "localhost",
				"--root", fmt.Sprintf("%s/assets/domain/localhost/", currentDir),
			},
			requests: []testRequest{
				{
					url:      "http://localhost:%d",
					status:   http.StatusOK,
					response: "localhost",
				},
				{
					url:      "http://127.0.0.1:%d",
					status:   http.StatusOK,
					response: "localhost",
				},
			},
		},
		{
			label: "Test Https Server",
			args: []string{
				"--domain", "localhost",
				"--root", fmt.Sprintf("%s/assets/domain/localhost/", currentDir),
				"--cert", fmt.Sprintf("%s/assets/cert/server_cert.crt", currentDir),
				"--key", fmt.Sprintf("%s/assets/cert/server_cert.key", currentDir),
			},
			cert: fmt.Sprintf("%s/assets/cert/server_cert.crt", currentDir),
			key:  fmt.Sprintf("%s/assets/cert/server_cert.key", currentDir),
			ca:   fmt.Sprintf("%s/assets/cert/root_ca.crt", currentDir),
			requests: []testRequest{
				{
					url:      "https://localhost:%d",
					status:   http.StatusOK,
					response: "localhost",
				},
			},
		},
		{
			label: "Test Gzip",
			args: []string{
				"--domain", "localhost",
				"--root", fmt.Sprintf("%s/assets/domain/localhost/", currentDir),
			},
			requests: []testRequest{
				{
					url:      "http://localhost:%d/gzip.html",
					status:   http.StatusOK,
					response: "gzip",
				},
			},
		},
		{
			label: "Test Single Page Routing",
			args: []string{
				"--domain", "localhost",
				"--root", fmt.Sprintf("%s/assets/domain/localhost/", currentDir),
				"--mode", "history",
			},
			requests: []testRequest{
				{
					url:      "http://localhost:%d/",
					status:   http.StatusOK,
					response: "localhost",
				},
				{
					url:      "http://localhost:%d/a/b/c",
					status:   http.StatusOK,
					response: "localhost",
				},
			},
		},
		{
			label: "Test Proxy",
			args: []string{
				"--domain", "localhost",
				"--root", fmt.Sprintf("%s/assets/domain/localhost/", currentDir),
				"--proxy", "/proxy/gen_204:http://connectivitycheck.gstatic.com/generate_204",
				"--proxy", "/proxy/another/a/b/c/gen_204:http://connectivitycheck.gstatic.com/generate_204",
			},
			requests: []testRequest{
				{
					url:      "http://localhost:%d/proxy/gen_204",
					status:   http.StatusNoContent,
					response: "",
				},
				{
					url:      "http://localhost:%d/proxy/another/a/b/c/gen_204",
					status:   http.StatusNoContent,
					response: "",
				},
			},
		},
		{
			label: "Test Multiple Domains",
			args: []string{
				"--domain", "example.net",
				"--root", fmt.Sprintf("%s/assets/domain/example.net/", currentDir),
				"--domain", "localhost",
				"--root", fmt.Sprintf("%s/assets/domain/localhost/", currentDir),
			},
			requests: []testRequest{
				{
					url:      "http://127.0.0.1:%d",
					status:   http.StatusOK,
					response: "example.net",
				},
				{
					url:      "http://localhost:%d",
					status:   http.StatusOK,
					response: "localhost",
				},
			},
		},
		{
			label:               "Test SelfSigned Https Server",
			enableSelfSignedSSL: true,
			args: []string{
				"--domain", "localhost",
				"--root", fmt.Sprintf("%s/assets/domain/localhost/", currentDir),
			},
			requests: []testRequest{
				{
					// 首次访问，CA证书还没生成，所以客户端无法受信，会出错
					url:      "https://localhost:%d",
					allowErr: true,
				},
				{
					url:      "https://localhost:%d",
					status:   http.StatusOK,
					response: "localhost",
				},
			},
		},
	}

	for _, c := range testCaseList {
		c.Run(t)
	}
}

func TestMain(m *testing.M) {
	// 获取当前文件夹
	currentDir, _ = os.Getwd()
	selfSignedCertDir := getSelfSignedCertDir()
	if _, err := os.Stat(selfSignedCertDir); err == nil {
		os.RemoveAll(selfSignedCertDir)
	}

	os.Exit(m.Run())
}
