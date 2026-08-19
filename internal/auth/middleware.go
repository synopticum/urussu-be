package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// Middleware returns an HTTP middleware that requires a valid Bearer token
// for every request except those under publicPaths (e.g. the auth endpoints
// themselves). On success the user ID and role are injected into the request
// context; on failure it writes a 401 in the grpc-gateway error shape
// ({"code": 16, ...}, 16 = UNAUTHENTICATED) so clients parse one error format.
func Middleware(secret string, log *slog.Logger, publicPaths ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublic(r.URL.Path, publicPaths) {
				next.ServeHTTP(w, r)
				return
			}

			tokenString, ok := bearerToken(r)
			if !ok {
				writeUnauthorized(w, "missing bearer token")
				return
			}

			userID, role, err := Parse(secret, tokenString)
			if err != nil {
				if IsExpired(err) {
					writeUnauthorized(w, "token expired")
				} else {
					writeUnauthorized(w, "invalid token")
				}
				return
			}

			log.DebugContext(r.Context(), "authenticated request",
				slog.String("user_id", userID), slog.String("role", role))

			next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), userID, role)))
		})
	}
}

func isPublic(path string, publicPaths []string) bool {
	for _, prefix := range publicPaths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func bearerToken(r *http.Request) (string, bool) {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

// gatewayError mirrors the JSON body grpc-gateway writes for its own errors.
type gatewayError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	// Encoding cannot fail for this fixed shape.
	_ = json.NewEncoder(w).Encode(gatewayError{Code: 16, Message: message})
}
