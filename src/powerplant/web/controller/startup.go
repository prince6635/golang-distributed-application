package controller

import "net/http"

var webSocket = newWebsocketController()

func Initialize() {
	registerRoutes()
	registerFileServers()
}

func registerRoutes() {
	http.HandleFunc("/ws", webSocket.handleMessage)
}

func registerFileServers() {
	http.Handle("/public/",
		http.FileServer(http.Dir("/Users/ziwang/Uber/sync/ziwangdev1.dev.uber.com/home/uber/gocode/src/github.com/golang-distributed-application/src/powerplant/web/assets")))

	// Need to move files in /public/lib/ to node_modules
	// http.Handle("/public/lib/",
	// 	http.StripPrefix("/public/lib/", http.FileServer(http.Dir("node_modules"))))
}
