package src

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	if domain != nil {
		// 检查代理配置
		isProxy := handleProxy(domain, &w, r)
		if isProxy {
			return
		}
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
		// 检查代理配置
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

	if os.IsNotExist(err) || config.Root == "" {
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

func handleProxy(domain *DomainConfig, w *http.ResponseWriter, r *http.Request) (isProxy bool) {
	proxies := domain.Proxy
	skipTlsVerify := domain.SkipTLSVerify
	isProxy = false
	if proxies == nil {
		return
	}
	path := r.URL.Path
	var proxyConfig *DomainProxy
	for i := 0; i < len(*proxies); i++ {
		if strings.Index(path, (*proxies)[i].Url) == 0 {
			proxyConfig = &(*proxies)[i]
		}
	}
	if proxyConfig != nil {
		isProxy = true
		if proxyConfig.Instance == nil {
			proxyConfig.Instance = &httputil.ReverseProxy{
				Director: func(r *http.Request) {
					path := r.URL.Path
					pathIndex := strings.Index(path, proxyConfig.Url)
					fullUrl := proxyConfig.Proxy + path[pathIndex+len(proxyConfig.Url):]
					parsedUrl, err := url.Parse(fullUrl)
					fmt.Printf("%s %s --> %s\n", domain.Domain, path, fullUrl)
					if err == nil {
						r.URL.Scheme = parsedUrl.Scheme
						r.URL.Host = parsedUrl.Host
						r.Host = parsedUrl.Host
						r.URL.Path = parsedUrl.Path
					}
				},
			}
			if skipTlsVerify {
				proxyConfig.Instance.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
			}
		}
		if strings.ToLower(r.Header.Get("connection")) == "upgrade" || strings.ToLower(r.Header.Get("upgrade")) == "websocket" {
			// 需要代理 websocket
			pathIndex := strings.Index(path, proxyConfig.Url)
			fullUrl := proxyConfig.Proxy + path[pathIndex+len(proxyConfig.Url):]
			fullUrl = strings.Replace(fullUrl, "http", "ws", 1)
			fmt.Printf("%s %s --> %s\n", domain.Domain, path, fullUrl)
			handleWebSocketProxy(fullUrl, domain.SkipTLSVerify, *w, r)
		} else {
			proxyConfig.Instance.ServeHTTP(*w, r)
		}
	}
	return
}

func handleWebSocketProxy(destURLStr string, skipTlsVerify bool, w http.ResponseWriter, r *http.Request) {
	destURL, err := url.Parse(destURLStr)
	if err != nil {
		http.Error(w, "Invalid destination URL", http.StatusInternalServerError)
		return
	}

	// WebSocket握手
	h, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := h.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	destReq := r.Clone(r.Context())
	destReq.Host = destURL.Host
	destReq.URL.Path = destURL.Path
	destReq.URL.RawPath = destURL.RawPath
	destReq.RequestURI = destURL.RawPath

	destPort := destURL.Port()

	var destConn net.Conn
	if destURL.Scheme == "wss" {
		if destPort == "" {
			destPort = "443"
		}
		// 建立TLS连接
		destConn, err = tls.Dial("tcp", fmt.Sprintf("%s:%s", destURL.Host, destPort), &tls.Config{
			InsecureSkipVerify: skipTlsVerify, // 根据需要设置此项，跳过证书验证
		})
	} else {
		if destPort == "" {
			destPort = "80"
		}
		// 建立TCP连接
		destConn, err = net.Dial("tcp", destURL.Host)
	}

	if err != nil {
		http.Error(w, "Failed to connect to destination server", http.StatusInternalServerError)
		return
	}
	defer destConn.Close()

	// 将客户端的请求写入目标服务器连接
	err = destReq.Write(destConn)
	if err != nil {
		http.Error(w, "Failed to write request to destination server", http.StatusInternalServerError)
		return
	}

	// 开始转发消息
	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}
