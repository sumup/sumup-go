// OAuth 2.0 Authorization Code flow with SumUp
//
// This example walks your through the steps necessary to implement
// OAuth 2.0 (<https://oauth.net/>) in case you are building a software
// for other people to use.
//
// To get started, you will need your client credentials.
// If you don't have any yet, you can create them in the
// [Developer Settings](https://me.sumup.com/en-us/settings/oauth2-applications).
//
// Your credentials need to be configured with the correct redirect URI,
// that's the URI the user will get redirected to once they authenticate
// and authorize your application. For development, you might want to
// use for example `http://localhost:8080/callback`. In production, you would
// redirect the user back to your host, e.g. `https://example.com/callback`.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/client"
	"github.com/sumup/sumup-go/merchants"
)

const StateCookieName = "oauth_state"
const PKCECookieName = "oauth_pkce"

func main() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURI := os.Getenv("REDIRECT_URI")

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     sumup.OAuth2Endpoint,
		// Scope is a mechanism in OAuth 2.0 to limit an application's access to a user's account.
		// You should always request the minimal set of scope that you need for your application to
		// work. In this example we use "email profile" scope which gives you access to user's
		// email address and their profile.
		Scopes: []string{"email profile"},
	}

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state, err := getRandomState()
		if err != nil {
			log.Fatalf("create random state: %v", err)
		}
		challenge := oauth2.GenerateVerifier()

		http.SetCookie(w, &http.Cookie{
			Name:  StateCookieName,
			Value: state,
			Path:  "/",
			// Set to true on production when running on https
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:  PKCECookieName,
			Value: challenge,
			Path:  "/",
			// Set to true on production when running on https
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// State is an opaque value used by the client to maintain state between the
		// request and callback. The authorization server includes this value when
		// redirecting the user agent back to the client.
		url := conf.AuthCodeURL(state, oauth2.S256ChallengeOption(challenge))
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		stateCookie, err := r.Cookie(StateCookieName)
		if err != nil {
			log.Fatalf("get oauth state cookie: %v", err)
		}

		pkceCookie, err := r.Cookie(PKCECookieName)
		if err != nil {
			log.Fatalf("get oauth pkce cookie: %v", err)
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if state != stateCookie.Value {
			log.Fatal("invalid oauth state")
		}

		token, err := conf.Exchange(r.Context(), code, oauth2.VerifierOption(pkceCookie.Value))
		if err != nil {
			log.Fatalf("retrieve token via code exchange: %v", err)
		}

		// Users might have access to multiple merchant accounts, the `merchant_code` parameter
		// returned in the callback is the merchant code of their default merchant account.
		// In production, you would want to let users pick which merchant they want to use
		// using the memberships API.
		defaultMerchantCode := r.URL.Query().Get("merchant_code")

		log.Printf("merchant code: %s", defaultMerchantCode)

		client := sumup.NewClient(client.WithAPIKey(token.AccessToken))

		merchant, err := client.Merchants.Get(r.Context(), defaultMerchantCode, merchants.GetMerchantParams{})
		if err != nil {
			log.Printf("get merchant information: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(merchant)
	})

	log.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getRandomState() (string, error) {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	if err != nil {
		return "", fmt.Errorf("generate new salt for user: %w", err)
	}
	return hex.EncodeToString(buf), err
}
