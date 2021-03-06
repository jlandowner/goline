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
	flag.StringVar(&idtoken, "idtoken", "", "ID Token https://developers.line.biz/ja/reference/line-login/#verify-id-token")
	flag.Parse()

	ctx := context.TODO()

	line := goline.NewClient(clientid, http.DefaultClient)

	p, err := line.VerifyIDToken(ctx, idtoken, "", "")
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("LINE User Name", p.Name)
}
