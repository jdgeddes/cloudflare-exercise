package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func certificateUpdate(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	bytes, _ := ioutil.ReadAll(request.Body)
	fmt.Println(string(bytes))
}

func main() {
	// setup HTTP router for handling customer and certificate requests
	r := httprouter.New()
	r.POST("/", certificateUpdate)

	// start external HTTP server
	http.ListenAndServe(":8081", r)
}
