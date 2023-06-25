package src

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type StaticServerHandler struct {
	Domains        []DomainConfig
	DefaultWWWRoot string
	NotFound       string
}

func (s *StaticServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var target string
	var code int
	domain := CurrentDomain(&s.Domains, r.Host)
	if domain != nil && domain.Root != "" {
		fmt.Printf("%s %s\n", domain.Domain, r.URL.Path)
		target, code = getSatisfiedFile(&findFileConfig{
			Root:     domain.Root,
			Path:     r.URL.Path,
			NotFound: s.NotFound,
		})
		if domain.Mode == "history" {
			filePath := path.Join(domain.Root, r.URL.Path)
			_, err := os.Stat(filePath)
			if err != nil && (filepath.Ext(filePath) == "" || filepath.Ext(filePath) == "html") {
				target, code = getSatisfiedFile(&findFileConfig{
					Root:     domain.Root,
					Path:     "index.html",
					NotFound: s.NotFound,
				})
			}
		}
	} else if s.DefaultWWWRoot != "" {
		fmt.Printf("%s %s\n", "default", r.URL.Path)
		target, code = getSatisfiedFile(&findFileConfig{
			Root:     s.DefaultWWWRoot,
			Path:     r.URL.Path,
			NotFound: s.NotFound,
		})
	}
	if code == 200 {
		// 判断使用本地gzip文件的情况
		if !strings.HasSuffix(r.URL.Path, ".gz") && strings.HasSuffix(target, ".gz") {
			sendFile(&w, target, 200)
			return
		}
		http.ServeFile(w, r, target)
	} else if code == 404 {
		if domain != nil && domain.NotFound != "" {
			sendFile(&w, domain.NotFound, 404)
		} else if s.NotFound != "" {
			sendFile(&w, s.NotFound, 404)
		} else {
			http.NotFound(w, r)
		}
	} else {
		w.WriteHeader(code)
		w.Write([]byte{})
	}
}

type findFileConfig struct {
	Root     string
	Path     string
	NotFound string
}

func getSatisfiedFile(config *findFileConfig) (target string, code int) {
	code = 200
	target = path.Join(config.Root, config.Path)
	info, err := os.Stat(target)

	if os.IsNotExist(err) {
		// 检查本地是否有gzip压缩文件
		_, err := os.Stat(fmt.Sprintf("%s.gz", target))
		if !(os.IsNotExist(err) || os.IsPermission(err)) {
			target = fmt.Sprintf("%s.gz", target)
			return
		}
		code = 404
		return
	}

	if os.IsPermission(err) {
		code = 403
		return
	}

	if info.IsDir() {
		target, code = getSatisfiedFile(&findFileConfig{
			Root: config.Root,
			Path: path.Join(config.Path, "index.html"),
		})
		return
	}

	checkGzipFileExist(&target)
	return
}

func checkGzipFileExist(path *string) {
	_, err := os.Stat(fmt.Sprintf("%s.gz", *path))
	if os.IsNotExist(err) {
		return
	}
	if os.IsPermission(err) {
		return
	}
	*path = fmt.Sprintf("%s.gz", *path)
}

func sendFile(w *http.ResponseWriter, file string, code int) {
	stream, err := os.ReadFile(file)
	if err == nil {
		if strings.HasSuffix(file, ".gz") {
			_, name := filepath.Split(file)
			exts := strings.Split(name, ".")
			ext := exts[len(exts)-2]
			contentType := mime.TypeByExtension(fmt.Sprintf(".%s", ext))
			if contentType != "" {
				(*w).Header().Set("content-type", contentType)
			}
			(*w).Header().Set("vary", "accept-encoding")
			(*w).Header().Set("content-encoding", "gzip")
		}
		(*w).WriteHeader(code)
		(*w).Write(stream)
	} else {
		(*w).WriteHeader(404)
	}
}
