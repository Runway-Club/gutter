// Package loginwithgoogle renders Google's official "Continue with Google"
// button via the Google Identity Services (GSI) JavaScript library and
// surfaces the resulting JWT to Go code as a typed credential.
//
// It lives under github.com/Runway-Club/gutter/community/ because it is a
// vendor-specific widget — Google branding, an OAuth client ID, and a
// runtime dependency on accounts.google.com. The core gutter widget
// catalog stays vendor-neutral; reusable third-party integrations like
// this one live alongside it under community/.
//
// Setup:
//
//  1. Create an OAuth 2.0 Client ID at
//     https://console.cloud.google.com/apis/credentials (Web application).
//  2. Add every origin you'll serve from (http://localhost:8080,
//     https://yourapp.example, …) to "Authorized JavaScript origins".
//  3. Construct loginwithgoogle.Button{ClientID: "...", OnCredential: ...}.
//
// The widget loads https://accounts.google.com/gsi/client on first mount
// (once per page), initializes GSI with ClientID, and lets Google paint
// its branded button into the hosting <div>. When the user signs in,
// OnCredential fires with a parsed [Credential].
//
// Trust: GSI delivers the JWT over a TLS channel from accounts.google.com.
// For production, send Credential.Token to your backend and re-verify
// against https://www.googleapis.com/oauth2/v3/certs before trusting any
// claim. The client-side parse done here is convenience-only.
package loginwithgoogle

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Runway-Club/gutter"
)

// GLogoDataURL is the four-color Google "G" mark as an inline SVG data: URL.
// Useful for app-side branding that mirrors what Google's own renderButton
// would paint — e.g. a custom-styled trigger that opens One Tap, or a
// placeholder while the GSI script loads.
//
// The string is URL-encoded (single quotes, %23 for #) so it can be dropped
// into an <img src> or background-image without further escaping.
const GLogoDataURL = "data:image/svg+xml;utf8," +
	"<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 48 48'>" +
	"<path fill='%23FFC107' d='M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12s5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24s8.955,20,20,20s20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z'/>" +
	"<path fill='%23FF3D00' d='M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z'/>" +
	"<path fill='%234CAF50' d='M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z'/>" +
	"<path fill='%231976D2' d='M43.611,20.083H42V20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571l0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z'/>" +
	"</svg>"

// Credential is the parsed payload of a successful sign-in. Token is the
// raw JWT — forward it to your backend for verification with Google's
// public keys before trusting any field on this struct.
type Credential struct {
	Token         string // raw JWT (the "credential" field GSI delivers)
	Sub           string // Google user ID, stable per (account, client_id)
	Email         string
	EmailVerified bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string // URL to the user's profile photo
	Issuer        string // typically "https://accounts.google.com"
	Audience      string // your client_id
	Expiry        int64  // unix seconds
}

// Button is the StatefulWidget that renders Google's official sign-in
// button. ClientID is required; OnCredential fires on success, OnError
// fires when the GSI script can't load or the JWT is malformed.
//
//	loginwithgoogle.Button{
//	    ClientID: "1234.apps.googleusercontent.com",
//	    Text:     "continue_with",
//	    OnCredential: func(c loginwithgoogle.Credential) {
//	        log.Printf("signed in as %s (%s)", c.Name, c.Email)
//	    },
//	}
//
// Visual options (Theme/Size/Text/Shape/Width) are forwarded directly to
// google.accounts.id.renderButton; see Google's docs for the accepted
// values. Empty values fall through to the GSI defaults
// (Theme="outline", Size="large", Text="continue_with").
type Button struct {
	ClientID     string
	OnCredential func(Credential)
	OnError      func(string)

	Theme string // "outline" | "filled_blue" | "filled_black"
	Size  string // "large" | "medium" | "small"
	Text  string // "signin_with" | "signup_with" | "continue_with" | "signin"
	Shape string // "rectangular" | "pill" | "circle" | "square"
	Width float64
}

func (b Button) CreateState() gutter.State { return &buttonState{} }

type buttonState struct {
	gutter.StateObject
	cleanup func()
}

func (s *buttonState) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.Widget().(Button)
	if w.ClientID == "" {
		return hostDiv{placeholder: "loginwithgoogle.Button: ClientID is empty"}
	}
	return hostDiv{
		onMount: func(node any) {
			if s.cleanup != nil {
				return
			}
			s.cleanup = renderGoogleSignIn(node, w, func(cred Credential) {
				if w.OnCredential != nil {
					w.OnCredential(cred)
				}
			}, func(err string) {
				if w.OnError != nil {
					w.OnError(err)
				}
			})
		},
	}
}

func (s *buttonState) Dispose() {
	if s.cleanup != nil {
		s.cleanup()
		s.cleanup = nil
	}
}

// hostDiv is the anchor element GSI paints the button into. Plain
// HostWidget so we can attach OnMount without depending on widgets
// package internals.
type hostDiv struct {
	onMount     func(node any)
	placeholder string
}

func (h hostDiv) Host() *gutter.Host {
	host := &gutter.Host{
		Tag:     "div",
		Style:   map[string]string{"display": "inline-block"},
		OnMount: h.onMount,
	}
	if h.placeholder != "" {
		host.Text = h.placeholder
		host.Style["padding"] = "10px 16px"
		host.Style["border"] = "1px dashed #c5c5c7"
		host.Style["border-radius"] = "8px"
		host.Style["color"] = "#888"
		host.Style["font-size"] = "13px"
	}
	return host
}

// parseJWT decodes the middle (payload) segment of a 3-part JWT and pulls
// the standard Google ID-token claims into a Credential. The signature is
// NOT verified — see the package doc for the trust model.
func parseJWT(jwt string) (Credential, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return Credential{}, fmt.Errorf("malformed JWT: expected 3 segments, got %d", len(parts))
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Credential{}, fmt.Errorf("decode payload: %w", err)
	}
	var claims struct {
		Iss           string `json:"iss"`
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Exp           int64  `json:"exp"`
	}
	if err := json.Unmarshal(data, &claims); err != nil {
		return Credential{}, fmt.Errorf("parse claims: %w", err)
	}
	return Credential{
		Token:         jwt,
		Sub:           claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		GivenName:     claims.GivenName,
		FamilyName:    claims.FamilyName,
		Picture:       claims.Picture,
		Issuer:        claims.Iss,
		Audience:      claims.Aud,
		Expiry:        claims.Exp,
	}, nil
}
