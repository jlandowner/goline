package main

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
