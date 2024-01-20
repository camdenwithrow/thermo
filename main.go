package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	targetPort, outputPort := getFlags()
	// proxyPort := 4321
	// targetPort := 8000
	targetURL := fmt.Sprintf("http://localhost:%s", targetPort)
	go monitorServer(targetURL)
	startProxyServer(targetURL, outputPort)
}

func getFlags() (string, string) {
	var port string
	var output string
	flag.StringVar(&port, "port", "8080", "port to listen on")
	flag.StringVar(&port, "p", "8080", "port to listen on (shorthand)")
	flag.StringVar(&output, "output", "8081", "port to forward to")
	flag.StringVar(&output, "o", "8081", "port to forward to (shorthand)")
	flag.Parse()
	return port, output
}

func startProxyServer(target string, port string) {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	fmt.Printf("Starting proxy server of %s at http://localhost:%s\n", target, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func monitorServer(target string) {
	serverIsUp := false
	for {
		_, err := http.Get(target)
		if err != nil {
			fmt.Println("Server is down :(")
			serverIsUp = false
		} else if !serverIsUp {
			// push websocket
			serverIsUp = true
			fmt.Println("Server is back up bb")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
