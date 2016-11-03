package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/arcticfoxnv/oolong/oauth"
)

type DummyOAuthClient struct {
}

func (c *DummyOAuthClient) GetAuthorizeURL() string {
	return "http://example.com/authorize"
}

func (c *DummyOAuthClient) GetAccessToken(code string) (string, error) {
	if code == "failnow" {
		return "", errors.New("Failed to get token")
	}
	return code, nil
}

func TestGetLocalIPAddress(t *testing.T) {
	ip := GetLocalIPAddress()
	if ip == "" {
		t.Fail()
	}

	if ip == "127.0.0.1" {
		t.Fail()
	}

	if ip == "::1" {
		t.Fail()
	}
}

func TestClientLoginHandler(t *testing.T) {
	resp := httptest.NewRecorder()
	path := "/start"

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		t.FailNow()
	}

	oauthClient = oauth.NewOAuthClient("test.id", "test.secret", "http://example.com/authorize")
	ClientLoginHandler(resp, req)

	if resp.Code != http.StatusFound {
		t.Fail()
	}

	if resp.Header().Get("Location") != oauthClient.GetAuthorizeURL() {
		t.Fail()
	}

}

func TestAuthorizeHandler(t *testing.T) {
	resp := httptest.NewRecorder()
	path := "/authorize"
	param := make(url.Values)
	param["code"] = []string{"123456"}

	req, err := http.NewRequest(http.MethodGet, path+param.Encode(), nil)
	if err != nil {
		t.FailNow()
	}

	done := make(chan int)
	oauthClient = &DummyOAuthClient{}
	go AuthorizeHandler(resp, req, done)
	<-done

	// TODO: Test valid token is written
}

func TestAuthorizeHandlerFailed(t *testing.T) {
	resp := httptest.NewRecorder()
	path := "/authorize"
	param := make(url.Values)
	param["code"] = []string{"failnow"}

	req, err := http.NewRequest(http.MethodGet, path+param.Encode(), nil)
	if err != nil {
		t.FailNow()
	}

	done := make(chan int)
	oauthClient = &DummyOAuthClient{}
	go AuthorizeHandler(resp, req, done)
	<-done

	// TODO: Test no token is written
}
