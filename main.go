package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Proxy is url port pair
type Proxy struct {
	RawURL   string
	ProxyURL string
}

// SSL used for https
type SSL struct {
	Cert string
	Key  string
}

// ProxyConfig Config
type ProxyConfig struct {
	ProxyList []Proxy
	SSL       *SSL
	Default   string
}

var path = flag.String("config", "", "reverse proxy config file path with yaml format")
var config *ProxyConfig

func main() {
	config = &ProxyConfig{}
	flag.Parse()
	if path == nil || len(*path) == 0 {
		panic(errors.New("Must specify config file path"))
	}
	f, err := os.Open(*path)
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		panic(err)
	}
	if len(config.Default) == 0 {
		panic(errors.New("config file must set defaulport"))
	}
	http.HandleFunc("/", ServeHTTP)
	if config.SSL != nil && len(config.SSL.Cert) > 0 && len(config.SSL.Key) > 0 {
		go http.ListenAndServeTLS(":443", config.SSL.Cert, config.SSL.Key, nil)
	}
	http.ListenAndServe(":80", nil)
}

// ServeHTTP handle http/https request
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer CatchPanic()
	proxyurl := config.Default
	for _, proxy := range config.ProxyList {
		if strings.HasPrefix(r.URL.Hostname(), proxy.RawURL) {
			proxyurl = proxy.ProxyURL
			break
		}
	}
	remote, err := url.Parse(fmt.Sprintf("http://%s/", proxyurl))
	if err != nil {
		panic(err)
	}
	proxy := NewSingleHostReverseProxy(remote, proxyurl)
	proxy.ServeHTTP(w, r)
}

// CatchPanic catch panic
func CatchPanic() {
	if err := recover(); err != nil {
		println(err)
	}
}

// NewSingleHostReverseProxy Copy From httputil.NewSingleHostReverseProxy And Pass host
func NewSingleHostReverseProxy(target *url.URL, host string) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.Host = host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
