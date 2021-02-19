package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	line "github.com/jlandowner/go-line-authorizer"
)

const (
	port int = 3000
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Get User name
	name := r.Header.Get("LINEDisplayName")

	log.Println("hello,", name)
	w.Write([]byte("hello," + name))
}

func main() {
	clientid := flag.String("clientid", "", "LINE Channel ID https://developers.line.biz/ja/reference/line-login/#verify-id-token")
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/", helloHandler)

	// Setup authorizer
	lauth := line.NewLINEAuthorizer(*clientid)

	// Use VerifyIDTokenMiddleware
	router.Use(lauth.VerifyIDTokenMiddleware)

	// Or Use VerifyAccessTokenMiddleware
	// router.Use(lauth.VerifyAccessTokenMiddleware)

	log.Println(http.ListenAndServe(":3000", router))
}
