package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	upgrader = newUpgrader()
)

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnOpen(func(c *websocket.Conn) {
		// echo
		fmt.Println("OnOpen:", c.RemoteAddr().String())
		c.WriteMessage(websocket.TextMessage, []byte("Hi there!"))
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		fmt.Println("OnMessage:", messageType, string(data))
		c.WriteMessage(messageType, data)
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		fmt.Println("OnClose:", c.RemoteAddr().String(), err)
	})
	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Upgraded:", conn.RemoteAddr().String())
}

func main() {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", onWebsocket)
	engine := nbhttp.NewEngine(nbhttp.Config{
		// Network:                 "tcp",
		Addrs: []string{"localhost:8780"},
		// MaxLoad:                 1000000,
		// ReleaseWebsocketPayload: true,
		Handler: mux,
	})

	mux.Handle("GET /debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("GET /debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("GET /debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("GET /debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("GET /debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}
