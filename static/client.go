package static

import (
	"crypto/tls"
	"io"
	"net/http"
)

func Get(url string) (content string, status int, err error) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	content = string(body)
	return
}
