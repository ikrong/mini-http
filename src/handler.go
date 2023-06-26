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
			Root: domain.Root,
			Path: r.URL.Path,
		})
		if domain.Mode == "history" {
			// 先判断路径下是否有文件
			target, code = getSatisfiedFile(&findFileConfig{
				Root: domain.Root,
				Path: r.URL.Path,
			})
			// 如果没有文件，并且请求html，则返回index.html
			if code == 404 && (filepath.Ext(target) == "" || filepath.Ext(target) == "html") {
				target, code = getSatisfiedFile(&findFileConfig{
					Root: domain.Root,
					Path: "index.html",
				})
			}
		}
	} else if s.DefaultWWWRoot != "" {
		fmt.Printf("%s %s\n", "default", r.URL.Path)
		target, code = getSatisfiedFile(&findFileConfig{
			Root: s.DefaultWWWRoot,
			Path: r.URL.Path,
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
	Root string
	Path string
}

func getSatisfiedFile(config *findFileConfig) (target string, code int) {
	code = 200
	target = path.Join(config.Root, config.Path)

	// 优先检查本地是否有gzip压缩文件
	_, gzipErr := os.Stat(fmt.Sprintf("%s.gz", target))
	if !(os.IsNotExist(gzipErr) || os.IsPermission(gzipErr)) {
		target = fmt.Sprintf("%s.gz", target)
		return
	}

	info, err := os.Stat(target)

	if os.IsNotExist(err) {
		code = 404
		target = ""
		return
	}

	if os.IsPermission(err) {
		code = 403
		target = ""
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
