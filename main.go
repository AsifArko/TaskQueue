package main

import (
	"gitlab.com/pub-sub/router"
	"net/http"
)

func main() {

	// Initialize the router
	router.InitRouter()

	// Listen and serve
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}

}
