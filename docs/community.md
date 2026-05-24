---
title: Community packages
nav_order: 9
---

# Community packages
{: .no_toc }

The `community/` tier holds reusable widgets that intentionally don't belong in the core widget catalog — typically because they depend on a specific vendor's library (Google, Facebook, Stripe) or a specific brand's design language. The core `widgets/` package stays vendor-neutral; opt into the community ones explicitly.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## What's there

- [`community/login_with_google`](#loginwithgoogle) — Google Identity Services button.

---

## `loginwithgoogle.Button`
{: #loginwithgoogle }

Real "Continue with Google" via Google Identity Services. Loads `https://accounts.google.com/gsi/client` once per page, calls `google.accounts.id.initialize/renderButton`, bridges the callback into Go via `js.FuncOf`, and parses the returned JWT into a typed `Credential`.

### Setup

1. Create an OAuth 2.0 Client ID at [Google Cloud Console](https://console.cloud.google.com/apis/credentials) (Web application).
2. Add every origin you'll serve from (`http://localhost:8080`, `https://yourapp.example`, …) to **Authorized JavaScript origins**.
3. Paste the client ID into your widget construction.

### Signature

```go
type Credential struct {
    Token         string // raw JWT
    Sub           string // stable Google user ID
    Email         string
    EmailVerified bool
    Name          string
    GivenName     string
    FamilyName    string
    Picture       string
    Issuer        string
    Audience      string
    Expiry        int64 // unix seconds
}

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
```

`Theme`, `Size`, `Text`, `Shape`, `Width` are forwarded directly to `google.accounts.id.renderButton`. Empty values fall back to GSI defaults (`outline / large / continue_with`).

### Usage

```go
import loginwithgoogle "github.com/Runway-Club/gutter/community/login_with_google"

loginwithgoogle.Button{
    ClientID: "1234.apps.googleusercontent.com",
    Text:     "continue_with",
    OnCredential: func(c loginwithgoogle.Credential) {
        log.Printf("signed in as %s (%s)", c.Name, c.Email)
        // Send c.Token to your backend for verification.
    },
    OnError: func(err string) { log.Println("login error:", err) },
}
```

When `ClientID` is empty, the widget renders an inert placeholder so the rest of the page still works during setup.

### Trust model

The JWT is delivered over TLS from `accounts.google.com`, but the client-side parse in this package does NOT verify the signature. For production code, send `Credential.Token` to your backend and re-verify against [Google's public keys](https://www.googleapis.com/oauth2/v3/certs) before trusting any claim.

### Bonus: `loginwithgoogle.GLogoDataURL`

An inline-SVG `data:` URL of the four-color Google "G" mark — useful for app-side branding that mirrors what Google's own button paints. Drop into `widgets.Image{Src: …}`.

---

## Adding a community package

Drop a new package under `community/<name>/`. It can import `gutter` and optionally `widgets` (e.g. to reuse `widgets.Image` for branding assets). The dependency direction stays one-way: nothing in `widgets/` or `gutter/` imports `community/`.

Style notes:

- Keep the public surface small and specific to the vendor.
- Use `_wasm.go` / `_stub.go` if you need `syscall/js`.
- Document the trust model when the integration crosses an authentication or payment boundary.
