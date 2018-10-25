// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Sample helloworld is a basic App Engine flexible app.
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

var (
	hueOauthConfig *oauth2.Config
)
var (
	oauthStateString = "pseudo-random"
)

func init() {
	var appid = os.Getenv("HUE_APPID")
	hueOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("CALLBACK_URL"),
		ClientID:     os.Getenv("HUE_CLIENT_ID"),
		ClientSecret: os.Getenv("HUE_CLIENT_SECRET"),
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://api.meethue.com/oauth2/auth?appid=%s&deviceid=%s&devicename=browser", appid, appid),
			TokenURL: fmt.Sprintf("https://api.meethue.com/oauth2/token"),
		},
	}
}

func main() {
	// AppEngine health check
	http.HandleFunc("/_ah/health", healthCheckHandler)

	// Oauth2 client routes
	http.HandleFunc("/login", handleHueLogin)
	http.HandleFunc("/hue_callback_url", handleHueCallback)

	// Debugger application
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/clip.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "clip.html")
	})

	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v2/") || strings.HasPrefix(r.URL.Path, "/bridge/") {
		handleRequestAndRedirect(w, r)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var htmlIndex = `<html>
<body>
	<a href="/login">Hue Log In</a>
</body>
</html>`
	fmt.Fprintf(w, htmlIndex)
}

func handleHueLogin(w http.ResponseWriter, r *http.Request) {
	url := hueOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleHueCallback(w http.ResponseWriter, r *http.Request) {
	content, err := getToken(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("clip.html?accessToken=%s", content.AccessToken), http.StatusTemporaryRedirect)
}

func getToken(state string, code string) (*oauth2.Token, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := hueOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	url, _ := url.Parse(target)
	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host
	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	req.URL.Host = "api.meethue.com"
	req.URL.Scheme = "https"
	log.Printf("Proxy %s %s", req.Method, req.URL.String())
	serveReverseProxy(req.URL.String(), res, req)
}
