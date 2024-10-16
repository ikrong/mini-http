package static

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"mini-http/log"
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
	serverConfig ServerConfig
}

func (s *StaticServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var target string
	var code int
	domain := s.serverConfig.CurrentDomain(r.Host)
	// 检查自动代理配置
	domain.readAutoProxyConfig(r)
	// 检查代理配置
	isProxy := handleProxy(domain, &w, r)
	if isProxy {
		return
	}
	log.Info("%s %s", domain.label(), r.URL.Path)
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
	if code == 200 {
		// 判断使用本地gzip文件的情况
		if !strings.HasSuffix(r.URL.Path, ".gz") && strings.HasSuffix(target, ".gz") {
			sendFile(&w, target, 200)
			return
		}
		http.ServeFile(w, r, target)
	} else if code == 404 {
		if domain.NotFound != "" {
			sendFile(&w, domain.NotFound, 404)
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
	isProxy = false
	if proxies == nil {
		return
	}
	path := r.URL.Path
	var proxyConfig *DomainProxy
	for i := 0; i < len(*proxies); i++ {
		if strings.Index(path, (*proxies)[i].Url) == 0 {
			proxyConfig = &(*proxies)[i]
			break
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
					log.Info("%s %s --> %s", domain.label(), path, fullUrl)
					if err == nil {
						r.URL.Scheme = parsedUrl.Scheme
						r.URL.Host = parsedUrl.Host
						r.Host = parsedUrl.Host
						r.URL.Path = parsedUrl.Path
					}
				},
				ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
					log.Error("%s %s --> %s", domain.label(), req.URL.Path, err.Error())
				},
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
		}
		if strings.ToLower(r.Header.Get("connection")) == "upgrade" || strings.ToLower(r.Header.Get("upgrade")) == "websocket" {
			// 需要代理 websocket
			pathIndex := strings.Index(path, proxyConfig.Url)
			fullUrl := proxyConfig.Proxy + path[pathIndex+len(proxyConfig.Url):]
			fullUrl = strings.Replace(fullUrl, "http", "ws", 1)
			log.Info("%s %s --> %s", domain.label(), path, fullUrl)
			err := handleWebSocketProxy(fullUrl, *w, r)
			if err != nil {
				log.Error("%s %s", domain.label(), err)
			}
		} else {
			proxyConfig.Instance.ServeHTTP(*w, r)
		}
	}
	return
}

func handleWebSocketProxy(destURLStr string, w http.ResponseWriter, r *http.Request) error {
	destURL, err := url.Parse(destURLStr)
	if err != nil {
		return err
	}

	// WebSocket握手
	h, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("WebSocket upgrade failed: %s", destURLStr)
	}
	clientConn, _, err := h.Hijack()
	if err != nil {
		return err
	}
	defer clientConn.Close()

	destReq := r.Clone(r.Context())
	destReq.Host = destURL.Host
	destReq.URL.Path = destURL.Path
	destReq.URL.RawPath = destURL.RawPath
	destReq.RequestURI = destURL.RawPath

	var destConn net.Conn
	if destURL.Scheme == "wss" {
		wssUrl := destURL.Host
		if destURL.Port() == "" {
			wssUrl = fmt.Sprintf("%s:443", destURL.Host)
		}
		// 建立TLS连接
		destConn, err = tls.Dial("tcp", wssUrl, &tls.Config{
			InsecureSkipVerify: true, // 根据需要设置此项，跳过证书验证
		})
	} else {
		// 建立TCP连接
		destConn, err = net.Dial("tcp", destURL.Host)
	}

	if err != nil {
		return err
	}
	defer destConn.Close()

	// 将客户端的请求写入目标服务器连接
	err = destReq.Write(destConn)
	if err != nil {
		return err
	}

	// 开始转发消息
	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)

	return nil
}
