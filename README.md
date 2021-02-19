# LINE Login for Golang
LINE Login package for Go and authorizer for http server

## Support API

- verify-id-token
  https://developers.line.biz/ja/reference/line-login/#verify-id-token

- verify-access-token
  https://developers.line.biz/ja/reference/line-login/#verify-access-token

- get-user-profile
  https://developers.line.biz/ja/reference/line-login/#get-user-profile


## Install
```sh
go get "github.com/jlandowner/go-line-authorizer"
```

## Example(verify-id-token)

### Use http Middleware

```go
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

### Use API directly

Also you can directly use LINE Login API functions implemented in this package.

```go
import (
	"context"
	line "github.com/jlandowner/go-line-authorizer"
)

func GetLINEUserNameByIDToken(clientid, idtoken string) (string, error) {
	ctx := context.TODO()

	p, err := line.VerifyIDToken(ctx, clientid, idtoken, "")
	if err != nil {
		return "", err
	}
	return p.Name, nil
}
```