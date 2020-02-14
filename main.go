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
	URL  string
	Port int
}

// SSL used for https
type SSL struct {
	Cert string
	Key  string
}

// ProxyConfig Config
type ProxyConfig struct {
	ProxyList   []Proxy
	SSL         *SSL
	DefaultPort int
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
	if config.DefaultPort == 0 {
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
	port := config.DefaultPort
	for _, proxy := range config.ProxyList {
		if strings.HasPrefix(r.URL.Hostname(), proxy.URL) {
			port = proxy.Port
			break
		}
	}
	remote, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}

// CatchPanic catch panic
func CatchPanic() {
	if err := recover(); err != nil {
		println(err)
	}
}
