package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/gotokatsuya/yahoo/yconnect/v2/attribute"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     os.Getenv("YCONNECT_CLIENT_ID"),
		ClientSecret: os.Getenv("YCONNECT_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.login.yahoo.co.jp/yconnect/v2/authorization",
			TokenURL: "https://auth.login.yahoo.co.jp/yconnect/v2/token",
		},
		RedirectURL: os.Getenv("YCONNECT_CALLBACK_URL"),
		Scopes:      []string{"openid", "profile", "email", "address"},
	}

	oauthStateString = "thisshouldberandom"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html><body>
		Logged in with <a href="/login">Yahoo! Japan</a>
		</body></html>
		`))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	option := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("display", "popup"),
	}
	url := oauthConf.AuthCodeURL(oauthStateString, option...)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleLoginCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	state := r.URL.Query().Get("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		fmt.Printf(" oauthConf.Exchange() failed with '%v'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	attributeClient := attribute.NewClient(http.DefaultClient)
	attributeRequest, err := attributeClient.NewRequest(&attribute.RequestBody{
		AccessToken: token.AccessToken,
	})
	if err != nil {
		fmt.Printf("attributelient.NewRequest() failed with '%v'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	attributeResponse, err := attributeClient.Do(ctx, attributeRequest)
	if err != nil {
		fmt.Printf("attributeClient.Do() failed with '%v'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(attributeResponse)
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/login/callback", handleLoginCallback)

	fmt.Print("Started running on http://localhost:3000\n")
	fmt.Println(http.ListenAndServe(":3000", nil))
}
