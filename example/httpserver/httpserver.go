package main

import (
	"errors"
	"flag"
	"log"
	"net/http"

	"github.com/go-logr/zapr"
	"github.com/gorilla/mux"
	"github.com/jlandowner/goline"
	"go.uber.org/zap"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Get User name
	name := r.Header.Get(goline.HeaderKeyLINEDisplayName)

	log.Println("hello,", name)
	w.Write([]byte("hello," + name))
}

func main() {
	clientid := flag.String("clientid", "", "LINE Channel ID https://developers.line.biz/ja/reference/line-login/#verify-id-token")
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/", helloHandler)

	// Setup logr
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log := zapr.NewLogger(zapLog)

	// Setup Client
	lineClient := goline.NewClient(*clientid, http.DefaultClient)

	// Setup Authorizer
	lineAuth := goline.NewAuthorizer(*clientid, lineClient, zapr.NewLogger(zapLog))

	// Use VerifyIDTokenMiddleware
	router.Use(lineAuth.VerifyIDTokenMiddleware)

	// Or Use VerifyAccessTokenMiddleware
	// router.Use(lauth.VerifyAccessTokenMiddleware)

	err = http.ListenAndServe(":3000", router)
	if !errors.Is(err, http.ErrServerClosed) {
		log.Error(err, "unexpected err")
	}
}
