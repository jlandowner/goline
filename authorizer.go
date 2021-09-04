package goline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
)

// Authorizer is a clientset of LINE Auth API
type Authorizer struct {
	clientID   string
	lineClient *Client
	log        logr.Logger
}

// NewAuthorizer return new Authorizer
func NewAuthorizer(clientid string, lineClient *Client, log logr.Logger) *Authorizer {
	return &Authorizer{clientID: clientid, lineClient: lineClient, log: log.WithName("goline.Authorizer")}
}

// VerifyIDTokenMiddleware is a middleware of http handler
// Obtain id token from authorization header and verify it upstream
// The authorized LINE user info is set in request headers "LINEUserID", "LINEDisplayName", "LINEPictureURL", "LINEEmail"
func (a *Authorizer) VerifyIDTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.log.WithName("VerifyAccessTokenMiddleware")
		ctx := context.TODO()

		bearerToken := r.Header.Get("Authorization")
		if bearerToken == "" {
			log.Error(errors.New("innvalid header"), "bearer token not found in authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		idToken, err := extractTokenFromBearerToken(bearerToken)
		if err != nil {
			log.Error(err, "failed to extract token form bearer")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		p, err := a.lineClient.VerifyIDToken(ctx, a.clientID, idToken, "", "")
		if err != nil || p == nil {
			log.Error(err, "failed to verify id token", "profile", p)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Add("LINEUserID", p.Sub)
		r.Header.Add("LINEDisplayName", p.Name)
		r.Header.Add("LINEPictureURL", p.Picutre)
		r.Header.Add("LINEEmail", p.Email)

		next.ServeHTTP(w, r)
	})
}

// VerifyAccessTokenMiddleware is a middleware of http handler
// Obtain access token from authorization header and verify it upstream
// The authorized LINE user info is set in request headers "LINEUserID", "LINEDisplayName", "LINEPictureURL", "LINEStatusMessage"
func (a *Authorizer) VerifyAccessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.log.WithName("VerifyAccessTokenMiddleware")
		ctx := context.TODO()

		bearerToken := r.Header.Get("Authorization")
		if bearerToken == "" {
			log.Error(errors.New("innvalid header"), "bearer token not found in authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		accessToken, err := extractTokenFromBearerToken(bearerToken)
		if err != nil {
			log.Error(err, "failed to extract token form bearer")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// first verify access token to check client ID
		res, err := a.lineClient.VerifyAccessToken(ctx, accessToken)
		if err != nil {
			log.Error(err, "failed to verify access token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if res.ClientID != a.clientID {
			log.Error(errors.New("invalid access token"), "client id not match as expected", "got", res.ClientID, "expected", a.clientID)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		p, err := a.lineClient.GetProfile(ctx, accessToken)
		if err != nil || p == nil {
			log.Error(err, "failed to get profile", "profile", p)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Add("LINEUserID", p.UserID)
		r.Header.Add("LINEDisplayName", p.DisplayName)
		r.Header.Add("LINEPictureURL", p.PictureURL)
		r.Header.Add("LINEStatusMessage", p.StatusMessage)

		next.ServeHTTP(w, r)
	})
}

func extractTokenFromBearerToken(bearerToken string) (string, error) {
	arr := strings.Split(bearerToken, "Bearer ")
	if len(arr) != 2 {
		return "", fmt.Errorf("Failed to get token")
	}
	return arr[1], nil
}
