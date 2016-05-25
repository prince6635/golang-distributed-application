package main

import (
	"net/http"

	"github.com/golang-distributed-application/src/powerplant/web/controller"
)

func main() {
	controller.Initialize()

	http.ListenAndServe(":3000", nil)
}
