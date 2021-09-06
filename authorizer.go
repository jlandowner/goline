package goline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
)

const (
	HeaderKeyLINEUserID        = "LINEUserID"
	HeaderKeyLINEDisplayName   = "LINEDisplayName"
	HeaderKeyLINEPictureURL    = "LINEPictureURL"
	HeaderKeyLINEEmail         = "LINEEmail"
	HeaderKeyLINEStatusMessage = "LINEStatusMessage"
)

// Authorizer is a clientset of LINE Auth API
type Authorizer struct {
	lineClient *Client
	log        logr.Logger
}

// NewAuthorizer return new Authorizer
func NewAuthorizer(clientid string, lineClient *Client, log logr.Logger) *Authorizer {
	return &Authorizer{lineClient: lineClient, log: log.WithName("goline.Authorizer")}
}

// VerifyIDTokenMiddleware is a middleware of http handler
// Obtain id token from authorization header and verify it upstream
// The authorized LINE user info is set in request headers "LINEUserID", "LINEDisplayName", "LINEPictureURL", "LINEEmail"
func (a *Authorizer) VerifyIDTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.log.WithName("VerifyAccessTokenMiddleware")
		ctx := context.TODO()

		authHeader := r.Header.Get(authHeader)
		if authHeader == "" {
			log.Error(errors.New("innvalid header"), "bearer token not found in authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		idToken, err := extractBearerToken(authHeader)
		if err != nil {
			log.Error(err, "failed to extract token form bearer")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		p, err := a.lineClient.VerifyIDToken(ctx, idToken, "", "")
		if err != nil || p == nil {
			log.Error(err, "failed to verify id token", "profile", p)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Add(HeaderKeyLINEUserID, p.Sub)
		r.Header.Add(HeaderKeyLINEDisplayName, p.Name)
		r.Header.Add(HeaderKeyLINEPictureURL, p.Picutre)
		r.Header.Add(HeaderKeyLINEEmail, p.Email)

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

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error(errors.New("innvalid header"), "bearer token not found in authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		accessToken, err := extractBearerToken(authHeader)
		if err != nil {
			log.Error(err, "failed to extract token form bearer")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// first verify access token to check client ID
		if _, err := a.lineClient.VerifyAccessToken(ctx, accessToken); err != nil {
			log.Error(err, "failed to verify access token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		p, err := a.lineClient.GetProfile(ctx, accessToken)
		if err != nil || p == nil {
			log.Error(err, "failed to get profile", "profile", p)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header.Add(HeaderKeyLINEUserID, p.UserID)
		r.Header.Add(HeaderKeyLINEDisplayName, p.DisplayName)
		r.Header.Add(HeaderKeyLINEPictureURL, p.PictureURL)
		r.Header.Add(HeaderKeyLINEStatusMessage, p.StatusMessage)

		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(authHeader string) (string, error) {
	arr := strings.Split(authHeader, "Bearer ")
	if len(arr) != 2 {
		return "", fmt.Errorf("not bearer")
	}
	return arr[1], nil
}
