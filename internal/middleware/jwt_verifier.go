package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/config"
	jwt "github.com/golang-jwt/jwt/v4"
)

// HS256Verifier verifies HS256 JWTs using the provided config
type HS256Verifier struct{}

func NewHS256Verifier() *HS256Verifier { return &HS256Verifier{} }

func (v *HS256Verifier) Verify(tokenString string, cfg *config.JWTConfig) (*JWTClaims, error) {
	if cfg == nil || cfg.Secret == "" {
		return nil, errors.New("jwt config/secret not set")
	}

	parsedToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, errors.New("invalid token signature")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// Extract times with leeway
	now := time.Now()
	leeway := cfg.Leeway

	// nbf
	if nbfVal, ok := claims["nbf"].(float64); ok {
		nbf := time.Unix(int64(nbfVal), 0)
		if nbf.After(now.Add(leeway)) {
			return nil, fmt.Errorf("token not yet valid: nbf=%v", nbf)
		}
	}
	// exp
	if expVal, ok := claims["exp"].(float64); ok {
		exp := time.Unix(int64(expVal), 0)
		if now.After(exp.Add(leeway)) {
			return nil, fmt.Errorf("token expired: exp=%v", exp)
		}
	}
	// iss
	if iss, ok := claims["iss"].(string); ok {
		if cfg.Issuer != "" && iss != cfg.Issuer {
			return nil, fmt.Errorf("issuer mismatch: got=%s want=%s", iss, cfg.Issuer)
		}
	}
	// aud (support string or array)
	if cfg.Audience != "" {
		if aud, ok := claims["aud"].(string); ok {
			if aud != cfg.Audience {
				return nil, fmt.Errorf("audience mismatch: got=%s want=%s", aud, cfg.Audience)
			}
		} else if audArr, ok := claims["aud"].([]interface{}); ok {
			match := false
			for _, v := range audArr {
				if s, ok := v.(string); ok && s == cfg.Audience {
					match = true
					break
				}
			}
			if !match {
				return nil, fmt.Errorf("audience mismatch: got=%v want=%s", audArr, cfg.Audience)
			}
		}
	}

	// Build JWTClaims
	out := &JWTClaims{Raw: map[string]any{}}
	if sub, ok := claims["sub"].(string); ok {
		out.Sub = sub
	}
	if iss, ok := claims["iss"].(string); ok {
		out.Iss = iss
	}
	if aud, ok := claims["aud"].(string); ok {
		out.Aud = aud
	}
	if exp, ok := claims["exp"].(float64); ok {
		out.Exp = time.Unix(int64(exp), 0)
	}
	if nbf, ok := claims["nbf"].(float64); ok {
		out.Nbf = time.Unix(int64(nbf), 0)
	}
	if iat, ok := claims["iat"].(float64); ok {
		out.Iat = time.Unix(int64(iat), 0)
	}
	for k, v := range claims {
		out.Raw[k] = v
	}
	return out, nil
}
