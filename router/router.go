package router

import (
	"gitlab.com/pub-sub/controller"
	"net/http"
)

func InitRouter() {

	psCtrl := controller.GetPubSubController()

	// Whenever the client do a GET on the index a client connection is established
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/home/hoods/go/src/gitlab.com/pub-sub/router/static")

	})

	// Websocket controller for handling the clients
	http.HandleFunc("/ws", psCtrl.WebSocketHandler)
}
