package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/DrC0ns0le/bind-api/config"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ClaimsContextKey contextKey = "claims"

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// jwksCache caches JWKS to avoid fetching on every request
type jwksCache struct {
	sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	jwksURL   string
}

var globalJWKSCache = &jwksCache{
	keys: make(map[string]*rsa.PublicKey),
}

// getPublicKey retrieves a public key by kid, fetching JWKS if needed
func (c *jwksCache) getPublicKey(kid string) (*rsa.PublicKey, error) {
	c.RLock()
	if time.Now().Before(c.expiresAt) {
		if key, ok := c.keys[kid]; ok {
			c.RUnlock()
			return key, nil
		}
	}
	c.RUnlock()

	// Fetch fresh JWKS
	if err := c.refresh(); err != nil {
		return nil, err
	}

	c.RLock()
	defer c.RUnlock()
	if key, ok := c.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("key with kid '%s' not found in JWKS", kid)
}

// refresh fetches the JWKS from the configured URL
func (c *jwksCache) refresh() error {
	c.Lock()
	defer c.Unlock()

	// Double-check after acquiring write lock
	if time.Now().Before(c.expiresAt) {
		return nil
	}

	jwksURL := config.GetEnv("JWKS_URL", "")
	if jwksURL == "" {
		return fmt.Errorf("JWKS_URL environment variable not set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Parse all RSA keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			continue
		}
		pubKey, err := jwkToRSAPublicKey(jwk)
		if err != nil {
			log.Printf("Warning: failed to parse JWK kid=%s: %v", jwk.Kid, err)
			continue
		}
		newKeys[jwk.Kid] = pubKey
	}

	c.keys = newKeys
	// Cache for 1 hour (JWKS doesn't change often)
	cacheDuration := config.GetEnv("JWKS_CACHE_DURATION", "3600")
	duration, err := time.ParseDuration(cacheDuration + "s")
	if err != nil {
		duration = time.Hour
	}
	c.expiresAt = time.Now().Add(duration)
	c.jwksURL = jwksURL

	log.Printf("JWKS refreshed: %d keys loaded from %s", len(newKeys), jwksURL)
	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}
	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// AuthMiddleware validates JWT tokens using JWKS from any OAuth2/OIDC provider
// Requires JWKS_URL environment variable (e.g., https://auth.example.com/realms/myrealm/protocol/openid-connect/certs)
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse token without validation first to get the kid
		unverifiedToken, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			http.Error(w, `{"error": "malformed token"}`, http.StatusUnauthorized)
			return
		}

		// Get kid from token header
		kid, ok := unverifiedToken.Header["kid"].(string)
		if !ok || kid == "" {
			http.Error(w, `{"error": "token missing kid header"}`, http.StatusUnauthorized)
			return
		}

		// Get public key from JWKS cache
		pubKey, err := globalJWKSCache.getPublicKey(kid)
		if err != nil {
			log.Printf("JWKS error: %v", err)
			http.Error(w, `{"error": "failed to validate token"}`, http.StatusUnauthorized)
			return
		}

		// Parse and validate token with the public key
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure token uses RSA signing (RS256, RS384, RS512)
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return pubKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		fmt.Println("Claims: ", claims)

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
