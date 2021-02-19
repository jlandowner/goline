package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// See https://developers.line.biz/ja/reference/line-login-v2/#get-user-profile
	urlGetUserProfile = "https://api.line.me/v2/profile"
	// See https://developers.line.biz/ja/reference/line-login/#verify-access-token
	urlVerifyAccessToken = "https://api.line.me/oauth2/v2.1/verify"
	// See https://developers.line.biz/ja/reference/line-login/#verify-id-token
	urlVerifyIDToken = "https://api.line.me/oauth2/v2.1/verify"
)

var (
	// ErrBadRequest 400 Bad Request リクエストに問題があります。リクエストパラメータとJSONの形式を確認してください。
	ErrBadRequest = errors.New("400 Bad Request")
	// ErrUnauthorized 401 Unauthorized Authorizationヘッダーを正しく送信していることを確認してください。
	ErrUnauthorized = errors.New("401 Unauthorized")
	// ErrForbidden 403 Forbidden APIを使用する権限がありません。ご契約中のプランやアカウントに付与されている権限を確認してください。
	ErrForbidden = errors.New("403 Forbidden")
	// ErrTooManyRequests 429 Too Many Requests リクエスト頻度をレート制限内に抑えてください。
	ErrTooManyRequests = errors.New("429 Too Many Requests")
	// ErrInternalServerError 500 Internal Server Error APIサーバーの一時的なエラーです。
	ErrInternalServerError = errors.New("500 Internal Server Error")
)

// LINEAuthorizer is a clientset of LINE Auth API
type LINEAuthorizer struct {
	clientID string
}

// NewLINEAuthorizer return new LINEAuthorizer
func NewLINEAuthorizer(clientid string) *LINEAuthorizer {
	return &LINEAuthorizer{clientID: clientid}
}

// IDTokenData is the response json struct of verify-id-token API
// See more -> https://developers.line.biz/ja/reference/line-login/#verify-id-token
type IDTokenData struct {
	Iss     string   `json:"iss"`
	Sub     string   `json:"sub"`
	Aud     string   `json:"aud"`
	Exp     string   `json:"exp"`
	Nonce   string   `json:"nonce,omitempty"`
	Amr     []string `json:"amr,omitempty"`
	Name    string   `json:"name,omitempty"`
	Picutre string   `json:"picture,omitempty"`
	Email   string   `json:"email,omitempty"`
}

// VerifyIDTokenMiddleware is a injecton middleware in http handler
// Obtain id token from authorization header and Verify id token to authorize
func (l *LINEAuthorizer) VerifyIDTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.TODO()

		bearerToken := r.Header.Get("Authorization")
		if bearerToken == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		idToken, err := getTokenFromBearerToken(bearerToken)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p, err := l.VerifyIDToken(ctx, idToken, "")
		if err != nil || p == nil {
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

func (l *LINEAuthorizer) VerifyIDToken(ctx context.Context, idToken, userid string) (*IDTokenData, error) {
	return VerifyIDToken(ctx, l.clientID, idToken, userid)
}

// VerifyIDToken is a function to call verify-id-token
func VerifyIDToken(ctx context.Context, clientid, idToken, userid string) (*IDTokenData, error) {
	// Check token paramater
	if idToken == "" {
		return nil, errors.New("ID Token not found")
	}

	// Prepare http request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlVerifyIDToken, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearerToken(idToken))
	req.URL.Query().Add("clientid", clientid)
	req.URL.Query().Add("nonce", strconv.Itoa(int(time.Now().UnixNano())))
	if userid != "" {
		req.URL.Query().Add("userid", userid)
	}

	// Do http request and get response body
	body, err := doRequestGetBody(req)
	if err != nil {
		return nil, err
	}

	return newIDTokenData(body)
}

func newIDTokenData(resBody []byte) (*IDTokenData, error) {
	p := &IDTokenData{}
	if err := json.Unmarshal(resBody, p); err != nil {
		return nil, err
	}
	return p, nil
}

// VerifyAccessTokenResponse is the response json struct of verify-access-token API
// See more -> https://developers.line.biz/ja/reference/line-login/#verify-access-token
type VerifyAccessTokenResponse struct {
	Scope     string `json:"scope"`
	ClientID  string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
}

// VerifyAccessTokenMiddleware is a injecton middleware in http handler
// Obtain access token from authorization header and Verify it to authorize
func (l *LINEAuthorizer) VerifyAccessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.TODO()

		bearerToken := r.Header.Get("Authorization")
		if bearerToken == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		accessToken, err := getTokenFromBearerToken(bearerToken)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		res, err := l.VerifyAccessToken(ctx, accessToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if res.ClientID != l.clientID {
			log.Printf("client id not match. get %s want %s", res.ClientID, l.clientID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p, err := GetProfile(ctx, accessToken)
		if err != nil || p == nil {
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

// VerifyAccessToken is a function to call verify-access-token API
func (l *LINEAuthorizer) VerifyAccessToken(ctx context.Context, accessToken string) (*VerifyAccessTokenResponse, error) {
	return VerifyAccessToken(ctx, accessToken)
}

// VerifyAccessToken is a function to call verify-access-token API
func VerifyAccessToken(ctx context.Context, accessToken string) (*VerifyAccessTokenResponse, error) {
	// Check token paramater
	if accessToken == "" {
		return nil, errors.New("Access Token not found")
	}

	// Prepare http request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlVerifyAccessToken, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Query().Add("access_token", accessToken)

	// Do http request and get response body
	body, err := doRequestGetBody(req)
	if err != nil {
		return nil, err
	}

	return newVerifyAccessTokenResponse(body)
}

func newVerifyAccessTokenResponse(resBody []byte) (*VerifyAccessTokenResponse, error) {
	p := &VerifyAccessTokenResponse{}
	if err := json.Unmarshal(resBody, p); err != nil {
		return nil, err
	}
	return p, nil
}

// LINEProfile is the response json struct of get-user-profile API
// See more -> https://developers.line.biz/ja/reference/line-login-v2/#get-profile-response
type LINEProfile struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

// GetProfile is a function to call get-user-profile API
func GetProfile(ctx context.Context, accessToken string) (*LINEProfile, error) {
	// Check token paramater
	if accessToken == "" {
		return nil, errors.New("Access Token not found")
	}

	// Prepare http request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlGetUserProfile, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearerToken(accessToken))

	// Do http request and get response body
	body, err := doRequestGetBody(req)
	if err != nil {
		return nil, err
	}

	return newLINEProfile(body)
}

func newLINEProfile(resBody []byte) (*LINEProfile, error) {
	p := &LINEProfile{}
	if err := json.Unmarshal(resBody, p); err != nil {
		return nil, err
	}
	return p, nil
}

func doRequestGetBody(req *http.Request) ([]byte, error) {
	// Do http request
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Check Status Code
	if res.StatusCode != http.StatusOK {
		return nil, errByStatusCode(res.StatusCode)
	}

	// Read response body
	return ioutil.ReadAll(res.Body)
}

func errByStatusCode(statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusInternalServerError:
		return ErrInternalServerError
	default:
		return fmt.Errorf("Unknown status code %d", statusCode)
	}
}

func bearerToken(token string) string {
	return "Bearer " + token
}

func getTokenFromBearerToken(bearerToken string) (string, error) {
	arr := strings.Split(bearerToken, "Bearer ")
	if len(arr) != 2 {
		return "", fmt.Errorf("Failed to get token")
	}
	return arr[1], nil
}
