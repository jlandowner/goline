package main

import (
	"context"
	"fmt"

	line "github.com/jlandowner/go-line-authorizer"
)

func GetLINEUserNameByAccessToken(clientid, accesstoken string) (string, error) {
	ctx := context.TODO()

	if res, err := line.VerifyAccessToken(ctx, accesstoken); err != nil {
		return "", err
	} else if res.ClientID != clientid {
		return "", fmt.Errorf("client id not match")
	}

	p, err := line.GetProfile(ctx, accesstoken)
	if err != nil {
		return "", err
	}
	return p.DisplayName, nil
}
