// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package debugger

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
	// HueOauthConfig specifies how to use OAuth
	HueOauthConfig *oauth2.Config
)
var (
	oauthStateString = "pseudo-random"
)

// MakeConfig sets-up the OAuth config
func MakeConfig(callbackURL, appID, clientID, clientSecret, apiEndPoint string) *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  callbackURL,
		ClientID:     clientID,
		ClientSecret: clientID,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/auth?appid=%s&deviceid=%s&devicename=browser", apiEndPoint, appID, appID),
			TokenURL: fmt.Sprintf("%s/oauth2/token", apiEndPoint),
		},
	}
}

func init() {
	HueOauthConfig = MakeConfig(
		os.Getenv("CALLBACK_URL"),
		os.Getenv("HUE_APPID"),
		os.Getenv("HUE_CLIENT_ID"),
		os.Getenv("HUE_CLIENT_SECRET"),
		"https://api.meethue.com")
}

func main() {
	// AppEngine health check
	http.HandleFunc("/_ah/health", healthCheckHandler)

	// Oauth2 client routes
	http.HandleFunc("/login", HandleHueLogin)
	http.HandleFunc("/hue_callback_url", HandleHueCallback(redirectToRootWithToken))

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
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v2/") || strings.HasPrefix(r.URL.Path, "/bridge/") || strings.HasPrefix(r.URL.Path, "/connectionstatus") {
		HandleRequestAndRedirect(w, r)
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

// HandleHueLogin redirects to Hue Login
func HandleHueLogin(w http.ResponseWriter, r *http.Request) {
	url := HueOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func redirectToRootWithToken(token *oauth2.Token, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("clip.html?accessToken=%s", token.AccessToken), http.StatusTemporaryRedirect)
}

// HandleHueCallback redirects to
func HandleHueCallback(handler func(token *oauth2.Token, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := GetToken(r.FormValue("state"), r.FormValue("code"))
		if err != nil {
			fmt.Println(err.Error())
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		handler(content, w, r)
	}
}

// GetToken retrieves an OAuth token from the API
func GetToken(state string, code string) (*oauth2.Token, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := HueOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}

// ServeReverseProxy acts as a reverse proxy for a given url
func ServeReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
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

// HandleRequestAndRedirect Given a request send it to the appropriate url
func HandleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	req.URL.Host = "api.meethue.com"
	req.URL.Scheme = "https"
	log.Printf("Proxy %s %s", req.Method, req.URL.String())
	ServeReverseProxy(req.URL.String(), res, req)
}
