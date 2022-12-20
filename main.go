// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package debugger

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"golang.org/x/oauth2"
)

type Debugger struct {
	// HueOauthConfig specifies how to use OAuth
	HueOAuthConfig *oauth2.Config
	// APIEndpoint to use for proxying
	APIEndpoint string
	// Root, is the root url of the client application
	Root string
	// StateFn generator
	StateFn func() string
	// State validator
	ValidateState func(string) bool
}

type AuthVersion int

const (
	V1 = AuthVersion(1)
	V2 = AuthVersion(2)
)

// MakeConfig sets-up the OAuth config
func MakeConfig(callbackURL, appID, clientID, clientSecret, apiEndPoint string, version AuthVersion) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/v2/oauth2/authorize?appid=%s&deviceid=%s&devicename=browser", apiEndPoint, appID, appID),
		TokenURL: fmt.Sprintf("%s/v2/oauth2/token", apiEndPoint),
	}
	if version == V1 {
		endpoint = oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/auth?appid=%s&deviceid=%s&devicename=browser-v1", apiEndPoint, appID, appID),
			TokenURL: fmt.Sprintf("%s/oauth2/token", apiEndPoint),
		}
	}
	return &oauth2.Config{
		RedirectURL:  callbackURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{},
		Endpoint:     endpoint,
	}
}

// HandleHueLogin redirects to Hue Login
func (d Debugger) HandleHueLogin(w http.ResponseWriter, r *http.Request) {
	url := d.HueOAuthConfig.AuthCodeURL(d.StateFn())
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleHueCallback redirects to
func (d Debugger) HandleHueCallback(handler func(token *oauth2.Token, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.FormValue("code"))
		content, err := d.GetToken(r.FormValue("state"), r.FormValue("code"))
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s?error=%s", d.Root, url.QueryEscape(err.Error())), http.StatusTemporaryRedirect)
			return
		}
		handler(content, w, r)
	}
}

// GetToken retrieves an OAuth token from the API
func (d Debugger) GetToken(state string, code string) (*oauth2.Token, error) {
	if d.ValidateState != nil && !d.ValidateState(state) {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := d.HueOAuthConfig.Exchange(context.TODO(), code)
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
func (d Debugger) HandleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	target, err := url.Parse(d.APIEndpoint)
	if err != nil {
		log.Fatalf("Invalid APIEndpoint")
	}
	originalURL := req.URL.String()

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)
	// Update the RequestURI & headers to allow for SSL redirection
	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = target.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	log.Printf("Proxy %s %s to %s", req.Method, originalURL, req.URL.String())
	proxy.ServeHTTP(res, req)
}
