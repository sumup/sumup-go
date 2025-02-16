package sumup

import "golang.org/x/oauth2"

// OAuth2Endpoint is SumUp's OAuth 2.0 endpoint.
var OAuth2Endpoint = oauth2.Endpoint{
	AuthURL:  "https://api.sumup.com/authorize",
	TokenURL: "https://api.sumup.com/token",
}
