package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/jlandowner/goline"
)

func main() {
	var accesstoken, clientid string
	flag.StringVar(&accesstoken, "accesstoken", "", "LINE Access Token https://developers.line.biz/ja/reference/line-login/#verify-access-token")
	flag.StringVar(&clientid, "clientid", "", "LINE Channel ID https://developers.line.biz/ja/reference/line-login/#verify-id-token")
	flag.Parse()

	ctx := context.TODO()

	line := goline.Client{Client: http.DefaultClient}

	if res, err := line.VerifyAccessToken(ctx, accesstoken); err != nil {
		log.Fatalln(err)

	} else if res.ClientID != clientid {
		log.Fatalln("client id not match")
	}

	p, err := line.GetProfile(ctx, accesstoken)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("LINE User Name", p.DisplayName)
}
