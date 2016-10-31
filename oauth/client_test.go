package oauth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAuthorizeURL(t *testing.T) {
	client := NewOAuthClient("abc", "123", "http://example.com")
	authUrl := client.GetAuthorizeURL()
	if authUrl != "https://www.mytaglist.com/oauth2/authorize.aspx?client_id=abc&redirect_uri=http://example.com" {
		t.Fail()
	}
}

func TestGetAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token": "xyz"}`)
	}))
	defer ts.Close()

	urlAccessToken = ts.URL

	client := NewOAuthClient("abc", "123", "http://example.com")
	token, err := client.GetAccessToken("xxx")
	if err != nil {
		t.Fail()
	}
	if token != "xyz" {
		t.Fail()
	}
}

func TestGetAccessTokenBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `some non-json garbage`)
	}))
	defer ts.Close()

	urlAccessToken = ts.URL

	client := NewOAuthClient("abc", "123", "http://example.com")
	token, err := client.GetAccessToken("xxx")
	if err == nil {
		t.Fail()
	}
	if token != "" {
		t.Fail()
	}
}
