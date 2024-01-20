package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

	modifyResponse := modifyResponseWithOutputPort(port)
	proxy.ModifyResponse = modifyResponse

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/ws", handleWebSocket)

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

func modifyResponseWithOutputPort(port string) func(res *http.Response) error {
	return func(res *http.Response) error {
		return modifyResponse(port, res)
	}
}

func modifyResponse(port string, res *http.Response) error {
	contentTypes := strings.Split(res.Header.Get("Content-Type"), ";")
	isHTML := false
	for _, contentType := range contentTypes {
		if strings.TrimSpace(contentType) == "text/html" {
			isHTML = true
		}
	}

	if isHTML {
		htmlBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		err = res.Body.Close()
		if err != nil {
			return err
		}

		// Append WebSocket script to the end of the body
		html := string(htmlBytes)
		bodyClosingTag := "</body>"
		script := fmt.Sprintf(`<script>(function() { var ws = new WebSocket('ws://localhost:%s/ws'); ws.onmessage = function() { window.location.reload(); }; })();</script>`, port)

		pos := strings.LastIndex(html, bodyClosingTag)
		newHTML := htmlBytes
		if pos != -1 {
			newHTML = []byte(html[:pos] + script + html[pos:])
		}

		res.Body = io.NopCloser(bytes.NewReader(newHTML))
		res.ContentLength = int64(len(newHTML))
		res.Header.Set("Content-Length", strconv.Itoa(len(newHTML)))
	}
	return nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

var clients = make(map[*websocket.Conn]bool) // connected clients

// handleWebSocket handles incoming WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	clients[conn] = true

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			break
		}
	}

}
