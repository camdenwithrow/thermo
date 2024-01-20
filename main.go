package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// TODO: Setup CLI Flags
	proxyPort := 4321
	targetPort := 8000
	targetURL := fmt.Sprintf("http://localhost:%d", targetPort)
	startProxyServer(targetURL, proxyPort)
	go monitorServer(targetURL)
}

func startProxyServer(target string, port int) {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}

func monitorServer(target string) {
}
