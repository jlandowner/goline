# Goline - Simple LINE Login for Go

## Support API

- verify-id-token
  https://developers.line.biz/ja/reference/line-login/#verify-id-token

- verify-access-token
  https://developers.line.biz/ja/reference/line-login/#verify-access-token

- get-user-profile
  https://developers.line.biz/ja/reference/line-login/#get-user-profile


## Install
```sh
go get "github.com/jlandowner/goline"
```

### Example

call verify-id-token API

```go
package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/jlandowner/goline"
)

func main() {
	var clientid, idtoken string
	flag.StringVar(&clientid, "clientid", "", "LINE Channel ID https://developers.line.biz/ja/reference/line-login/#verify-id-token")
	flag.StringVar(&idtoken, "idtoken", "", "LINE Channel ID https://developers.line.biz/ja/reference/line-login/#verify-id-token")
	flag.Parse()

	ctx := context.TODO()

	line := goline.Client{Client: http.DefaultClient}

	p, err := line.VerifyIDToken(ctx, clientid, idtoken, "")
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("LINE User Name", p.Name)
}
```

### Use http Middleware

This package prepares http Middleware easy to integrate LINE Login in your http server.

Here is a example server

```go
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
	name := r.Header.Get("LINEDisplayName")

	log.Println("hello,", name)
	w.Write([]byte("hello," + name))
}

func main() {
	clientid := flag.String("clientid", "", "LINE Channel ID https://developers.line.biz/ja/reference/line-login/#verify-id-token")
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/hello", helloHandler)

	// Setup logr
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log := zapr.NewLogger(zapLog)

	// Setup Client
	lineClient := &goline.Client{Client: http.DefaultClient}

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
```

```sh
# Run server
$ go run main.go --clientid="LINE Channel ID" &

# Request without token
$ curl -i http://localhost:3000/hello
400 Bad Request

# Request with valid token
# This assume you have already got ID Token in your client apps (Webclient(JavaScript), Android, iOS or others)
$ idtoken="ID Token"
$ curl -i http://localhost:3000/hello -H "Authorization: Bearer $idtoken"
200 OK

hello, XXX
```
