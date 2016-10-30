package main

import (
    "fmt"
    "log"
    "math/rand"
    "net"
    "net/http"

    "github.com/arcticfoxnv/oolong/oauth"
    "github.com/arcticfoxnv/oolong/wirelesstag"
)

var oauthClient oauth.OAuthClient

// Helper function to find the local IP.
func GetLocalIPAddress() string {
    ifaces, _ := net.Interfaces()
    for _, i := range ifaces {
        addrs, _ := i.Addrs()
        for _, addr := range addrs {
            var ip net.IP
            switch v := addr.(type) {
            case *net.IPNet:
                    ip = v.IP
            case *net.IPAddr:
                    ip = v.IP
            }
            if ip.IsGlobalUnicast() {
                return ip.String()
            }
        }
    }
    return ""
}

// Redirects the user to the OAuth authorization page on www.mytaglist.com/www.wirelesstag.net
func ClientLoginHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, oauthClient.GetAuthorizeURL(), http.StatusFound)
}

// OAuth page sends user to here, which reads the code and exchanges it for an access token.
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    log.Printf("Got auth code %s from client\n", code)
    accessToken, err := oauthClient.GetAccessToken(code)
    if err != nil {
        log.Printf("Failed to exchange code for token: %s\n", err.Error())
        return
    }
    log.Printf("Got access token: %s\n", accessToken)
    tagClient = wirelesstag.NewClient(accessToken)

    // Create a new state, and store the access token
    oolongState = NewOolongState()
    oolongState.AccessToken = accessToken

    // Write the state to file, so future starts can skip the oauth
    oolongState.WriteToFile()
    clientReady <- 1
}

func StartHTTPServer(config *Config, startChan chan string) {
    ip := GetLocalIPAddress()
    port := config.HTTP.Port
    if port == 0 {
        port = rand.Intn(1000) + 10000
    }
    // Redirect URL will be http://<local ip>:<port>/authorize
    redirectURL := fmt.Sprintf("http://%s:%d/authorize", ip, port)
    oauthClient = oauth.NewOAuthClient(config.OAuth.ID, config.OAuth.Secret, redirectURL)

    http.HandleFunc("/start", ClientLoginHandler)
    http.HandleFunc("/authorize", AuthorizeHandler)

    // Start URL will be http://<local ip>:<port>/authorize
    // Neither URL needs to be reachable from the remote OAuth server.  The only client
    // to access either URL will be the user's browser.
    startChan <- fmt.Sprintf("http://%s:%d/start", ip, port)

    // We'll listen forever... on the first run.  Subsequent runs won't start the server at all.
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
