package middleware

import (
	"context"
	"net/http"
	"strings"
)

type AuthService interface {
	VerifyAccessToken(ctx context.Context, rawToken string) error
}

func WithAuth(authService AuthService) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := r.Header.Clone()
			authHeader := r.Header.Get("Authorization")

			const bearerPrefix = "Bearer "
			// if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
			if authHeader == "" {
				http.Error(w, "Missing access token", http.StatusUnauthorized)
				return
			}

			rawToken := strings.TrimPrefix(authHeader, bearerPrefix)

			err := authService.VerifyAccessToken(r.Context(), rawToken)
			if err != nil {
				// Если токен отсутствует, невалиден или просрочен — возвращаем статус 401 Unauthorized.
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			r.Header.Set("Authorization", headers.Get("Authorization"))

			next.ServeHTTP(w, r)
		})
	}
}
