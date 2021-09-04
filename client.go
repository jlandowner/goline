package goline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// Client is an http client access to LINE Login API
type Client struct {
	Client *http.Client
}

// IDTokenData is the response json struct of verify-id-token API.
// https://developers.line.biz/ja/reference/line-login/#verify-id-token
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

// VerifyIDToken is a function to call verify-id-token.
// UserID and Nonce can be empty when not use.
// https://developers.line.biz/ja/reference/line-login/#verify-id-token
func (c *Client) VerifyIDToken(ctx context.Context, clientid, idToken, userid, nonce string) (*IDTokenData, error) {
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
	req.URL.Query().Add("nonce", nonce)
	if userid != "" {
		req.URL.Query().Add("userid", userid)
	}

	// Do http request and get response body
	d := &IDTokenData{}
	if err := c.doRequestGetBody(req, d); err != nil {
		return nil, err
	}
	return d, nil
}

// VerifyAccessTokenResponse is the response json struct of verify-access-token API.
// https://developers.line.biz/ja/reference/line-login/#verify-access-token
type VerifyAccessTokenResponse struct {
	Scope     string `json:"scope"`
	ClientID  string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
}

// VerifyAccessToken is a function to call verify-access-token API
// https://developers.line.biz/ja/reference/line-login/#verify-access-token
func (c *Client) VerifyAccessToken(ctx context.Context, accessToken string) (*VerifyAccessTokenResponse, error) {
	// Check token paramater
	if accessToken == "" {
		return nil, errors.New("Access Token not found")
	}

	// Prepare http request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlVerifyAccessToken, nil)
	if err != nil {
		return nil, err
	}
	params := req.URL.Query()
	params.Add("access_token", accessToken)
	req.URL.RawQuery = params.Encode()

	// Do http request and get response body
	res := &VerifyAccessTokenResponse{}
	if err := c.doRequestGetBody(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// LINEProfile is the response json struct of get-user-profile API
// https://developers.line.biz/ja/reference/line-login-v2/#get-profile-response
type LINEProfile struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

// GetProfile is a function to call get-user-profile API
// https://developers.line.biz/ja/reference/line-login-v2/#get-profile-response
func (c *Client) GetProfile(ctx context.Context, accessToken string) (*LINEProfile, error) {
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
	p := &LINEProfile{}
	if err := c.doRequestGetBody(req, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Client) doRequestGetBody(req *http.Request, v interface{}) error {
	// Do http request
	res, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Check Status Code
	if res.StatusCode != http.StatusOK {
		return errByStatusCode(res.StatusCode)
	}

	// Read response body
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, v); err != nil {
		return err
	}
	return nil
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
