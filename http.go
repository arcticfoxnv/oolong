package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"

	"github.com/arcticfoxnv/oolong/oauth"
	"github.com/arcticfoxnv/oolong/state"
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
func AuthorizeHandler(w http.ResponseWriter, r *http.Request, state state.State, done chan int) {
	code := r.URL.Query().Get("code")
	log.Printf("Got auth code %s from client\n", code)
	accessToken, err := oauthClient.GetAccessToken(code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %s\n", err.Error())
		return
	}
	log.Printf("Got access token: %s\n", accessToken)

	// Store the access token
	state.SetAccessToken(accessToken)

	// Write the state to file
	state.Save()
	done <- 1
}

func StartHTTPServer(config *Config, urlChan chan string, done chan int) {
	var st state.State
	ip := GetLocalIPAddress()
	port := config.HTTP.Port
	if port == 0 {
		port = rand.Intn(1000) + 10000
	}
	// Redirect URL will be http://<local ip>:<port>/authorize
	redirectURL := fmt.Sprintf("http://%s:%d/authorize", ip, port)
	oauthClient = oauth.NewOAuthClient(config.OAuth.ID, config.OAuth.Secret, redirectURL)

	// Create a new state
	switch config.Backend {
	case "file":
		st = state.NewFileState(config.File.Filename)
	case "redis":
		st = state.NewRedisState(config.Redis.Host, config.Redis.Port, config.Redis.Key)
	}

	http.HandleFunc("/start", ClientLoginHandler)
	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		AuthorizeHandler(w, r, st, done)
	})

	// Start URL will be http://<local ip>:<port>/authorize
	// Neither URL needs to be reachable from the remote OAuth server.  The only client
	// to access either URL will be the user's browser.
	urlChan <- fmt.Sprintf("http://%s:%d/start", ip, port)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
