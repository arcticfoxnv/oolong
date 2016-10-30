package oauth

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
)

type OAuthClient interface {
    GetAuthorizeURL() string
    GetAccessToken(string) (string, error)
}

type oauthClient struct {
    clientId     string
    clientSecret string
    redirectUrl  string
}

func NewOAuthClient(id, secret, redirect string) OAuthClient {
    return &oauthClient{
        clientId: id,
        clientSecret: secret,
        redirectUrl: redirect,
    }
}

// GetAuthorizeURL returns the URL the browser should be redirected to.
func (c *oauthClient) GetAuthorizeURL() string {
    return fmt.Sprintf("https://www.mytaglist.com/oauth2/authorize.aspx?client_id=%s&redirect_uri=%s", c.clientId, c.redirectUrl)
}


// GetAccessToken exchanges the code from the user for an access token from the server
func (c *oauthClient) GetAccessToken(authCode string) (string, error) {

    // Make a request to the server, providing the client id+secret+code from user.
    postUrl := "https://www.mytaglist.com/oauth2/access_token.aspx"
    resp, err := http.PostForm(postUrl, url.Values{
        "client_id": {c.clientId},
        "client_secret": {c.clientSecret},
        "code": {authCode},
    })
    if err != nil { return "", err }
    defer resp.Body.Close()

    // Read the response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil { return "", err }

    // Decode the body
    decodedResponse := make(map[string]interface{}, 0)
    err = json.Unmarshal(body, &decodedResponse)
    if err != nil { return "", err }

    // Return the access token from the response
    return decodedResponse["access_token"].(string), nil
}
