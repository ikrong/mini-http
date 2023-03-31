package src

import (
	"net/http"
	"os"
	"path"
)

type StaticServerHandler struct {
	Domains        []DomainConfig
	DefaultWWWRoot string
}

func (s *StaticServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domain := CurrentDomain(&s.Domains, r.Host)
	if domain != nil && domain.Root != "" {
		if domain.Mode == "history" {
			filePath := path.Join(domain.Root, r.URL.Path)
			_, err := os.Stat(filePath)
			if err != nil {
				http.ServeFile(w, r, path.Join(domain.Root, "index.html"))
				return
			}
		}
		if !checkFileExist(domain.Root, r.URL.Path) {
			http.NotFound(w, r)
			return
		}
		http.FileServer(http.Dir(domain.Root)).ServeHTTP(w, r)
		return
	} else if s.DefaultWWWRoot != "" {
		if !checkFileExist(s.DefaultWWWRoot, r.URL.Path) {
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
