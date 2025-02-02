// Package auth includes authentication API functions
package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const tokenPath = "/identity/token"

// Token Response
type Token struct {
	Value      string `json:"access_token"`
	Expiration int    `json:"expires_in"`
	Created    int64
}

// Error response
type Error struct {
	Code    string `json:"errorCode"`
	Message string `json:"errorMessage"`
	Details string `json:"errorDetails"`
}

type GetTokenError struct {
	Code    int
	Message string
	Details string
}

var GetNow = func() time.Time {
	return time.Now()
}

var GetAuthURL = func(endpoint string) (string, error) {
	return url.JoinPath(endpoint, tokenPath)
}

func (e GetTokenError) Error() string {
	return fmt.Sprintf("cannot get token. error code: %v, message: %v, details: %v", e.Code, e.Message, e.Details)
}

func GetToken(endpoint, key string) (Token, error) {

	token := Token{}

	data := url.Values{}
	data.Add("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	data.Add("apikey", key)

	addr, _ := GetAuthURL(endpoint)

	resp, err := http.PostForm(addr, data)
	if err != nil {
		return token, fmt.Errorf("cannot POST data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		e := Error{}
		if err = json.NewDecoder(resp.Body).Decode(&e); err != nil {
			return token, fmt.Errorf("cannot decode error message with status %d from JSON: %w", resp.StatusCode, err)
		}
		return token, GetTokenError{resp.StatusCode, e.Message, e.Details}
	}

	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return token, fmt.Errorf("cannot decode result data from JSON: %w", err)
	}

	token.Created = GetNow().Unix()

	return token, nil
}
