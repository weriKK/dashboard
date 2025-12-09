package backend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// HMACConfig holds the shared secret for HMAC authentication
type HMACConfig struct {
	Secret string
}

var hmacAuth HMACConfig

// InitHMAC loads or generates the HMAC secret
func InitHMAC(secretPath string) error {
	// In production, load from secure storage (env var, vault, etc.)
	// For now, check if DASHBOARD_HMAC_SECRET env var is set
	secret := os.Getenv("DASHBOARD_HMAC_SECRET")
	if secret == "" {
		log.Println("WARNING: DASHBOARD_HMAC_SECRET not set. HMAC auth disabled.")
		log.Println("WARNING: You can generate a secret like this: head -c 32 /dev/urandom | base64")
		return nil
	}
	hmacAuth.Secret = secret
	log.Println("HMAC authentication enabled")
	return nil
}

// ComputeSignature generates HMAC-SHA256 signature for a request
func ComputeSignature(secret, method, path, timestamp, body string) string {
	payload := fmt.Sprintf("%s|%s|%s|%s", method, path, timestamp, body)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMACSignature checks if request signature is valid
// Signature header format: "timestamp:signature"
// Returns: (valid, errorMessage)
func VerifyHMACSignature(r *http.Request, requireSecret bool) (bool, string) {
	if hmacAuth.Secret == "" {
		if requireSecret {
			return false, "HMAC authentication not configured on server"
		}
		// HMAC not configured and not required, allow all
		return true, ""
	}

	// Get signature header
	authHeader := r.Header.Get("X-HMAC-Signature")
	if authHeader == "" {
		return false, "missing X-HMAC-Signature header"
	}

	parts := strings.Split(authHeader, ":")
	if len(parts) != 2 {
		return false, "invalid signature format (expected timestamp:signature)"
	}

	timestamp := parts[0]
	clientSig := parts[1]

	// Verify timestamp is recent (within 2 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false, "invalid timestamp"
	}

	requestTime := time.Unix(ts, 0)
	if time.Since(requestTime) > 2*time.Minute {
		return false, "request timestamp too old"
	}

	if time.Until(requestTime) > 30*time.Second {
		return false, "request timestamp in future"
	}

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return false, "failed to read request body"
	}
	// Restore body for handler
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Compute expected signature
	expectedSig := ComputeSignature(hmacAuth.Secret, r.Method, r.URL.Path, timestamp, string(bodyBytes))

	// Constant-time comparison
	if !hmac.Equal([]byte(clientSig), []byte(expectedSig)) {
		return false, "invalid signature"
	}

	return true, ""
}

// RequireHMACAuth middleware enforces HMAC signature verification
// requireSecret=true means endpoint requires HMAC to be configured
func RequireHMACAuth(handler http.Handler, requireSecret bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		valid, errMsg := VerifyHMACSignature(r, requireSecret)
		if !valid {
			log.Printf("HMAC auth failed: %s from %s", errMsg, r.RemoteAddr)
			http.Error(w, fmt.Sprintf("Unauthorized: %s", errMsg), http.StatusUnauthorized)
			return
		}

		log.Printf("HMAC auth passed for %s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}
