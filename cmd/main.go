package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	debugger "github.com/q42philips/hue-remote-api-debugger"

	"golang.org/x/oauth2"
)

var instance debugger.Debugger

func init() {
	instance.HueOAuthConfig = debugger.MakeConfig(
		os.Getenv("CALLBACK_URL"),
		os.Getenv("HUE_APPID"),
		os.Getenv("HUE_CLIENT_ID"),
		os.Getenv("HUE_CLIENT_SECRET"),
		"https://api.meethue.com/v2")
	instance.APIEndpoint = "https://api.meethue.com"
}

func main() {
	// AppEngine health check
	http.HandleFunc("/_ah/health", healthCheckHandler)

	// Oauth2 client routes
	http.HandleFunc("/login", instance.HandleHueLogin)
	http.HandleFunc("/hue_callback_url", instance.HandleHueCallback(redirectToRootWithToken))

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

func redirectToRootWithToken(token *oauth2.Token, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("clip.html?accessToken=%s", token.AccessToken), http.StatusTemporaryRedirect)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v2/") || strings.HasPrefix(r.URL.Path, "/bridge/") || strings.HasPrefix(r.URL.Path, "/connectionstatus") {
		instance.HandleRequestAndRedirect(w, r)
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
	fmt.Fprint(w, htmlIndex)
}
