package src

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

type StaticServerHandler struct {
	Domains        []DomainConfig
	DefaultWWWRoot string
	NotFound       string
}

func (s *StaticServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domain := CurrentDomain(&s.Domains, r.Host)
	if domain != nil && domain.Root != "" {
		fmt.Printf("%s %s\n", domain.Domain, r.URL.Path)
		if domain.Mode == "history" {
			filePath := path.Join(domain.Root, r.URL.Path)
			_, err := os.Stat(filePath)
			if err != nil {
				http.ServeFile(w, r, path.Join(domain.Root, "index.html"))
				return
			}
		}
		if !checkFileExist(domain.Root, r.URL.Path) {
			if domain.NotFound != "" && checkFileExist("/", domain.NotFound) {
				sendFile(&w, domain.NotFound)
				return
			}
			http.NotFound(w, r)
			return
		}
		http.FileServer(http.Dir(domain.Root)).ServeHTTP(w, r)
		return
	} else if s.DefaultWWWRoot != "" {
		fmt.Printf("%s %s\n", "default", r.URL.Path)
		if !checkFileExist(s.DefaultWWWRoot, r.URL.Path) {
			if s.NotFound != "" && checkFileExist("/", s.NotFound) {
				sendFile(&w, s.NotFound)
				return
			}
			http.NotFound(w, r)
			return
		}
		http.FileServer(http.Dir(s.DefaultWWWRoot)).ServeHTTP(w, r)
	}
}

func checkFileExist(wwwRoot string, urlpath string) (exist bool) {
	fileLocalPath := path.Join(wwwRoot, urlpath)
	info, err := os.Stat(fileLocalPath)

	if os.IsNotExist(err) {
		return
	}

	exist = true

	if info.IsDir() {
		_, err = os.Stat(path.Join(fileLocalPath, "index.html"))
		if os.IsNotExist(err) {
			exist = false
			return
		}
	}
	return
}

func sendFile(w *http.ResponseWriter, file string) {
	stream, err := ioutil.ReadFile(file)
	if err == nil {
		(*w).WriteHeader(404)
		(*w).Write(stream)
	}
}
