package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var tokenResp = `{
	"access_token": "API_Token",
	"refresh_token": "not_supported",
	"ims_user_id": 7176943,
	"token_type": "Bearer",
	"expires_in": 3600,
	"expiration": 1735159110,
	"scope": "ibm openid"
}`

var errorMethod = `{
	"errorCode": "BXNIM0060E",
	"errorMessage": "Web application exception 'javax.ws.rs.ClientErrorException' during request processing.",
	"errorDetails": "HTTP 405 Method Not Allowed"
}`

var errorBadReq = `{
	"errorCode": "BXNIM0109E",
	"errorMessage": "Property missing or empty.",
	"errorDetails": "Property 'grant_type' either missing or empty.",
	"context": {
		"requestId": "djZmMjg-9ff452dfe4d64789894538ca065712d1",
		"requestType": "incoming.Identity_Token",
		"userAgent": "httpyac",
		"url": "https://iam.cloud.ibm.com",
		"instanceId": "iamid-9-12-3788-ae8e184-6f5fdb8f98-v6f28",
		"threadId": "286ad0",
		"host": "iamid-9-12-3788-ae8e184-6f5fdb8f98-v6f28",
		"startTime": "26.12.2024 17:23:26:340 UTC",
		"endTime": "26.12.2024 17:23:26:345 UTC",
		"elapsedTime": "5",
		"locale": "en_US",
		"clusterName": "iam-id-prod-eu-gb-lon02"
	}
}`

func httpError(message, details string) string {

	e := Error{
		Code:    "test_error",
		Message: message,
		Details: details,
	}
	j, err := json.Marshal(&e)

	if err != nil {
		log.Fatalf("Cannot create JSON error: %v", err)
	}

	return string(j)
}

func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		err := r.ParseForm()

		if err != nil {
			w.WriteHeader(400)
			e := httpError("Cannot parse form", err.Error())
			fmt.Fprintln(w, e)
		}

		switch {
		case r.Method != "POST":
			w.WriteHeader(405)
			fmt.Fprintln(w, errorMethod)
		case r.Form.Get("grant_type") == "urn:ibm:params:oauth:grant-type:apikey":
			if k := r.Form.Get("apikey"); k == "GOOD_API_KEY" {
				w.WriteHeader(200)
				fmt.Fprintln(w, tokenResp)
			} else {
				w.WriteHeader(403)
				fmt.Fprintln(w, httpError("Wrong API Key", fmt.Sprintf("Given Key: %s", k)))
			}
		case r.Form.Get("grant_type") == "":
			w.WriteHeader(400)
			fmt.Fprintln(w, errorBadReq)
		default:
			w.WriteHeader(400)
			e := httpError("Wrong Input", r.PostForm.Encode())
			fmt.Fprintln(w, e)
		}

	}

	return httptest.NewServer(http.HandlerFunc(f))
}

func TestGetToken(t *testing.T) {

	testCases := []struct {
		name  string
		input string
		want  Token
		err   any
	}{
		{name: "GoodAPIKey", input: "GOOD_API_KEY", want: Token{Value: "API_Token", Expiration: 3600, Created: 1234}, err: nil},
		{name: "BadAPIKey", input: "BAD_API_KEY", want: Token{}, err: GetTokenError{403, "Wrong API Key", "Given Key: BAD_API_KEY"}},
	}

	server := mockServer()
	defer server.Close()

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			// Mocking clock
			GetNow = func() time.Time {
				return time.Unix(tt.want.Created, 0)
			}

			got, err := GetToken(server.URL, tt.input)

			if tt.err != nil && err != nil && !errors.Is(err, tt.err.(error)) {
				t.Errorf("Got error: '%v', Want error: '%v'", err, tt.err)
				return
			}

			if err != nil && tt.err == nil {
				t.Errorf("Got unexpected error: '%v'", err)
			}

			if tt.err != nil && err == nil {
				t.Errorf("Want error: '%v', but no error returned", tt.err)
			}

			if got != tt.want {
				t.Errorf("Got: '%+v', Want: '%+v'", got, tt.want)
			}

		})
	}
}
