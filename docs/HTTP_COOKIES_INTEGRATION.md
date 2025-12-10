# HTTP-Only Cookies Integration Guide

This document explains how downstream services can integrate with the AAA service's HTTP-only cookie authentication.

## Overview

The AAA service now supports HTTP-only cookies for browser-based authentication alongside the existing Bearer token method. This provides:

- **XSS Protection**: HTTP-only cookies cannot be accessed via JavaScript (`document.cookie`)
- **Automatic Transmission**: Browsers automatically include cookies in requests
- **Backward Compatibility**: Bearer token authentication continues to work

## Cookie Specification

| Attribute | `auth_token` | `refresh_token` |
|-----------|--------------|-----------------|
| **Purpose** | Access token for API requests | Token for refreshing expired access tokens |
| **HttpOnly** | `true` | `true` |
| **Secure** | `true` (prod/staging) | `true` (prod/staging) |
| **Path** | `/` | `/` |
| **MaxAge** | 3600 seconds (1 hour) | 604800 seconds (7 days) |
| **SameSite** | `Lax` (or `None` for cross-origin) | `Lax` (or `None` for cross-origin) |
| **Domain** | Configurable via `AAA_COOKIE_DOMAIN` | Configurable via `AAA_COOKIE_DOMAIN` |

### Cross-Subdomain Cookie Sharing

To share cookies across subdomains (e.g., `aaa.kisanlink.in`, `farmers.kisanlink.in`), configure the cookie domain:

| Environment Variable | Default | Example |
|---------------------|---------|---------|
| `AAA_COOKIE_DOMAIN` | `""` (exact host only) | `.kisanlink.in` |

**Important:** The domain must start with a dot (`.`) to allow subdomain sharing.

```bash
# Production - share cookies across all *.kisanlink.in subdomains
AAA_COOKIE_DOMAIN=.kisanlink.in

# Development - leave empty (cookies bound to localhost)
# AAA_COOKIE_DOMAIN=  (or don't set)
```

### Environment-Based Security

The `Secure` flag is controlled by the `APP_ENV` environment variable:

| APP_ENV Value | Secure Flag |
|---------------|-------------|
| `production`, `prod`, `staging` | `true` |
| `development`, `dev`, others | `false` |

## Authentication Flow

### 1. Login

**Endpoint:** `POST /api/v1/auth/login`

On successful login, the AAA service:
1. Sets `auth_token` cookie with the access token
2. Sets `refresh_token` cookie with the refresh token
3. Returns tokens in JSON response body (backward compatible)

```bash
curl -X POST https://aaa.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user@example.com", "password": "secret"}'

# Response Headers:
# Set-Cookie: auth_token=eyJhbGc...; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=Lax
# Set-Cookie: refresh_token=eyJhbGc...; Path=/; Max-Age=604800; HttpOnly; Secure; SameSite=Lax

# Response Body (unchanged for backward compatibility):
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": { ... }
}
```

### 2. Token Extraction (Middleware)

The AAA middleware extracts tokens in this order:
1. **Authorization Header** (preferred for service-to-service)
2. **Cookie** (fallback for browser clients)

```go
// Middleware logic (internal/middleware/auth_middleware.go:70-78)
authz := c.GetHeader("Authorization")
if strings.HasPrefix(authz, "Bearer ") {
    token = strings.TrimPrefix(authz, "Bearer ")
} else if cookie, err := c.Request.Cookie("auth_token"); err == nil {
    token = cookie.Value
}
```

### 3. Token Refresh

**Endpoint:** `POST /api/v1/auth/refresh`

When the access token expires, use the refresh token to obtain new tokens:

```bash
curl -X POST https://aaa.example.com/api/v1/auth/refresh \
  -H "Cookie: refresh_token=eyJhbGc..."

# Both cookies are rotated with new tokens
```

### 4. Logout

**Endpoint:** `POST /api/v1/auth/logout`

Logout clears both cookies by setting `MaxAge=-1`:

```bash
curl -X POST https://aaa.example.com/api/v1/auth/logout \
  -H "Cookie: auth_token=eyJhbGc..."

# Response Headers:
# Set-Cookie: auth_token=; Path=/; Max-Age=-1; HttpOnly; Secure
# Set-Cookie: refresh_token=; Path=/; Max-Age=-1; HttpOnly; Secure
```

## Downstream Service Integration

### Option 1: Validate with AAA Service (Recommended)

Forward the token to AAA service for validation:

```go
package main

import (
    "net/http"

    pb "github.com/Kisanlink/aaa-service/pkg/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

func AuthMiddleware(aaaClient pb.AuthServiceClient) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header or cookie
            token := extractToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            // Validate with AAA service via gRPC
            ctx := r.Context()
            resp, err := aaaClient.ValidateToken(ctx, &pb.ValidateTokenRequest{
                Token: token,
            })
            if err != nil || !resp.Valid {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            // Token is valid, proceed
            next.ServeHTTP(w, r)
        })
    }
}

func extractToken(r *http.Request) string {
    // Try Authorization header first
    auth := r.Header.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }

    // Fallback to cookie
    if cookie, err := r.Cookie("auth_token"); err == nil {
        return cookie.Value
    }

    return ""
}
```

### Option 2: Local JWT Validation

If your service has the JWT public key, validate locally:

```go
package main

import (
    "net/http"

    "github.com/golang-jwt/jwt/v5"
)

func LocalAuthMiddleware(publicKey interface{}) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            // Parse and validate JWT
            parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
                return publicKey, nil
            })
            if err != nil || !parsed.Valid {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            // Extract claims and add to context
            claims := parsed.Claims.(jwt.MapClaims)
            ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
            ctx = context.WithValue(ctx, "claims", claims)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Option 3: Gin Framework Integration

For Gin-based downstream services:

```go
package middleware

import (
    "strings"

    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var token string

        // Try Authorization header first
        auth := c.GetHeader("Authorization")
        if strings.HasPrefix(auth, "Bearer ") {
            token = strings.TrimPrefix(auth, "Bearer ")
        } else {
            // Fallback to cookie
            token, _ = c.Cookie("auth_token")
        }

        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }

        // Validate token (via AAA gRPC or local JWT validation)
        // ...

        c.Set("token", token)
        c.Next()
    }
}
```

### Option 4: Browser Client (JavaScript)

For frontend applications using fetch:

```javascript
// Login
const response = await fetch('https://aaa.example.com/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  credentials: 'include',  // REQUIRED for cookies
  body: JSON.stringify({ username, password })
});

// Subsequent API calls - cookies are sent automatically
const data = await fetch('https://api.example.com/some-endpoint', {
  credentials: 'include'  // REQUIRED for cookies
});

// Logout
await fetch('https://aaa.example.com/api/v1/auth/logout', {
  method: 'POST',
  credentials: 'include'
});
```

**CORS Configuration Required:**

```go
// On your downstream service
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "https://your-frontend.com")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

## JWT Claims Structure

The access token JWT contains:

```json
{
  "sub": "USR00000001",
  "username": "user@example.com",
  "phone_number": "+919876543210",
  "exp": 1733745600,
  "iat": 1733742000,
  "user_context": {
    "roles": [
      {
        "role_id": "ROL00000001",
        "role_name": "admin",
        "service_name": "farmers-service"
      }
    ],
    "organizations": [
      {
        "id": "ORGN00000001",
        "name": "Kisanlink",
        "role": "member"
      }
    ],
    "groups": [
      {
        "id": "GRP00000001",
        "name": "Default Group"
      }
    ]
  }
}
```

### Extracting Organization IDs

The AAA middleware extracts organization IDs from `user_context.organizations[]`:

```go
// Extract organization IDs from JWT (internal/middleware/auth_middleware.go:132-171)
var organizationIDs []string
if userContext, exists := claims.Raw["user_context"]; exists {
    if userCtxMap, ok := userContext.(map[string]interface{}); ok {
        if orgsData, hasOrgs := userCtxMap["organizations"]; hasOrgs {
            if orgsSlice, ok := orgsData.([]interface{}); ok {
                for _, orgItem := range orgsSlice {
                    if orgMap, ok := orgItem.(map[string]interface{}); ok {
                        if orgID, ok := orgMap["id"].(string); ok && orgID != "" {
                            organizationIDs = append(organizationIDs, orgID)
                        }
                    }
                }
            }
        }
    }
}
c.Set("organization_ids", organizationIDs)
```

Downstream services can use these IDs for multi-tenant filtering.

## Proxy Configuration

If your downstream service sits behind a reverse proxy (nginx, envoy, etc.), ensure cookies are forwarded:

### Nginx

```nginx
location /api/ {
    proxy_pass http://downstream-service:8080;
    proxy_pass_header Set-Cookie;
    proxy_cookie_domain ~^(.*)$ $1;
    proxy_set_header Cookie $http_cookie;
}
```

### Envoy

```yaml
http_filters:
  - name: envoy.filters.http.router
route_config:
  virtual_hosts:
    - name: backend
      domains: ["*"]
      routes:
        - match: { prefix: "/api/" }
          route:
            cluster: downstream-service
          request_headers_to_add:
            - header:
                key: "Cookie"
                value: "%REQ(Cookie)%"
```

## Security Considerations

1. **Always use HTTPS in production** - Cookies with `Secure` flag require HTTPS
2. **SameSite=Lax** - Protects against CSRF for non-GET requests
3. **HttpOnly=true** - Prevents XSS attacks from stealing tokens
4. **Short access token TTL** - 1 hour limits exposure if compromised
5. **Token rotation on refresh** - Both tokens are replaced on refresh

## Troubleshooting

### Cookie not sent by browser

- Verify `credentials: 'include'` in fetch requests
- Check CORS `Access-Control-Allow-Credentials: true`
- Ensure same-site or correct SameSite policy

### Cookie not set after login

- Check browser dev tools Network tab for Set-Cookie header
- Verify response is over HTTPS (for Secure cookies)
- Check for cookie blocking extensions

### Token validation fails

- Verify token hasn't expired (check `exp` claim)
- Ensure JWT public key matches the private key used for signing
- Check clock skew between services

## Code References

| Component | File | Lines |
|-----------|------|-------|
| Cookie setting | `internal/handlers/auth/auth_handler.go` | 47-89 |
| Cookie clearing | `internal/handlers/auth/auth_handler.go` | 91-128 |
| Secure context check | `internal/handlers/auth/auth_handler.go` | 130-133 |
| Cross-origin detection | `internal/handlers/auth/auth_handler.go` | 136-149 |
| Cookie config | `internal/config/security.go` | 111-118, 202-206 |
| Token extraction | `internal/middleware/auth_middleware.go` | 70-78 |
| Org ID extraction | `internal/middleware/auth_middleware.go` | 132-171 |
